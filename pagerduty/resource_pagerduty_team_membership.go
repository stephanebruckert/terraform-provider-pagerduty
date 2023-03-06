package pagerduty

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/heimweh/go-pagerduty/pagerduty"
)

func resourcePagerDutyTeamMembership() *schema.Resource {
	return &schema.Resource{
		Create: resourcePagerDutyTeamMembershipCreate,
		Read:   resourcePagerDutyTeamMembershipRead,
		Update: resourcePagerDutyTeamMembershipUpdate,
		Delete: resourcePagerDutyTeamMembershipDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"team_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"role": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "manager",
				ValidateDiagFunc: validateValueDiagFunc([]string{
					"observer",
					"responder",
					"manager",
				}),
			},
		},
	}
}

func maxRetries() int {
	return 4
}

func retryDelayMs() int {
	return 500
}

func calculateDelay(retryCount int) time.Duration {
	return time.Duration(retryCount*retryDelayMs()) * time.Millisecond
}

func fetchPagerDutyTeamMembershipWithRetries(d *schema.ResourceData, meta interface{}, errCallback func(error, *schema.ResourceData) error, retryCount int, neededRole string) error {
	if retryCount >= maxRetries() {
		return nil
	}
	fmt.Println("2")

	if err := fetchPagerDutyTeamMembership(d, meta, errCallback); err != nil {
		return err
	}
	fetchedRole, userId, teamId := d.Get("role").(string), d.Get("user_id"), d.Get("team_id")
	if strings.Compare(neededRole, fetchedRole) == 0 {
		return nil
	}
	fmt.Printf("[DEBUG] Warning role '%s' fetched from PD is different from the role '%s' from config for user: %s from team: %s, retrying...\n", fetchedRole, neededRole, userId, teamId)

	retryCount++
	time.Sleep(calculateDelay(retryCount))
	return fetchPagerDutyTeamMembershipWithRetries(d, meta, errCallback, retryCount, neededRole)
}

func fetchPagerDutyTeamMembership(d *schema.ResourceData, meta interface{}, errCallback func(error, *schema.ResourceData) error) error {
	client, err := meta.(*Config).Client()
	if err != nil {
		return err
	}
	fmt.Println("4")

	userID, teamID := resourcePagerDutyParseColonCompoundID(d.Id())
	fmt.Printf("[DEBUG] Reading user: %s from team: %s\n", userID, teamID)
	return resource.Retry(2*time.Minute, func() *resource.RetryError {
		resp, _, err := client.Teams.GetMembers(teamID, &pagerduty.GetMembersOptions{})
		if err != nil {
			errResp := errCallback(err, d)
			if errResp != nil {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(errResp)
			}

			return nil
		}

		for _, member := range resp.Members {
			if member.User.ID == userID {
				d.Set("user_id", userID)
				d.Set("team_id", teamID)
				d.Set("role", member.Role)

				return nil
			}
		}

		fmt.Printf("[WARN] Removing %s since the user: %s is not a member of: %s\n", d.Id(), userID, teamID)
		d.SetId("")

		return nil
	})
}
func resourcePagerDutyTeamMembershipCreate(d *schema.ResourceData, meta interface{}) error {
	client, err := meta.(*Config).Client()
	if err != nil {
		return err
	}

	userID := d.Get("user_id").(string)
	teamID := d.Get("team_id").(string)
	role := d.Get("role").(string)

	fmt.Printf("[DEBUG] Adding user: %s to team: %s with role: %s\n", userID, teamID, role)

	retryErr := resource.Retry(2*time.Minute, func() *resource.RetryError {
		if _, err := client.Teams.AddUserWithRole(teamID, userID, role); err != nil {
			if isErrCode(err, 500) {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}

		return nil
	})
	if retryErr != nil {
		return retryErr
	}

	d.SetId(fmt.Sprintf("%s:%s", userID, teamID))
	fmt.Println("1")
	return fetchPagerDutyTeamMembershipWithRetries(d, meta, genError, 0, d.Get("role").(string))
}

func resourcePagerDutyTeamMembershipRead(d *schema.ResourceData, meta interface{}) error {
	fmt.Println("3")

	return fetchPagerDutyTeamMembership(d, meta, handleNotFoundError)
}

func resourcePagerDutyTeamMembershipUpdate(d *schema.ResourceData, meta interface{}) error {
	client, err := meta.(*Config).Client()
	if err != nil {
		return err
	}

	userID := d.Get("user_id").(string)
	teamID := d.Get("team_id").(string)
	role := d.Get("role").(string)

	fmt.Printf("[DEBUG] Updating user: %s to team: %s with role: %s\n", userID, teamID, role)

	// To update existing membership resource, We can use the same API as creating a new membership.
	retryErr := resource.Retry(2*time.Minute, func() *resource.RetryError {
		if _, err := client.Teams.AddUserWithRole(teamID, userID, role); err != nil {
			if isErrCode(err, 500) {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}

		return nil
	})
	if retryErr != nil {
		return retryErr
	}

	d.SetId(fmt.Sprintf("%s:%s", userID, teamID))

	return fetchPagerDutyTeamMembershipWithRetries(d, meta, genError, 0, d.Get("role").(string))
}

func resourcePagerDutyTeamMembershipDelete(d *schema.ResourceData, meta interface{}) error {
	client, err := meta.(*Config).Client()
	if err != nil {
		return err
	}

	userID, teamID := resourcePagerDutyParseColonCompoundID(d.Id())

	fmt.Printf("[DEBUG] Removing user: %s from team: %s\n", userID, teamID)

	// Extracting Escalation Policies ids where this team referenced
	epsAssociatedToUser, err := extractEPsAssociatedToUser(client, userID)
	if err != nil {
		return err
	}

	epsDissociatedFromTeam, err := dissociateEPsFromTeam(client, teamID, epsAssociatedToUser)
	if err != nil {
		return err
	}

	// Retrying to give other resources (such as escalation policies) to delete
	retryErr := resource.Retry(2*time.Minute, func() *resource.RetryError {
		fmt.Printf("this shit1 %s\n", userID)

		if _, err := client.Teams.RemoveUser(teamID, userID); err != nil {
			if isErrCode(err, 400) {
				fmt.Println("this shit2")
				return nil
			}

			return resource.NonRetryableError(err)
		}
		return nil
	})
	if retryErr != nil {
		time.Sleep(2 * time.Second)
		return retryErr
	}

	d.SetId("")

	err = associateEPsBackToTeam(client, teamID, epsDissociatedFromTeam)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	return nil
}

func buildEPsIdsList(l []*pagerduty.OnCall) []string {
	eps := []string{}
	for _, o := range l {
		if o.EscalationPolicy != nil {
			eps = append(eps, o.EscalationPolicy.ID)
		}
	}
	return unique(eps)
}

func extractEPsAssociatedToUser(c *pagerduty.Client, userID string) ([]string, error) {
	var oncalls []*pagerduty.OnCall
	retryErr := resource.Retry(10*time.Second, func() *resource.RetryError {
		resp, _, err := c.OnCall.List(&pagerduty.ListOnCallOptions{UserIds: []string{userID}})
		if err != nil {
			time.Sleep(2 * time.Second)
			return resource.RetryableError(err)
		}
		oncalls = resp.Oncalls
		return nil
	})
	if retryErr != nil {
		return nil, retryErr
	}
	epsAssociatedToUser := buildEPsIdsList(oncalls)
	return epsAssociatedToUser, nil
}

func dissociateEPsFromTeam(c *pagerduty.Client, teamID string, eps []string) ([]string, error) {
	epsDissociatedFromTeam := []string{}
	for _, ep := range eps {
		retryErr := resource.Retry(10*time.Second, func() *resource.RetryError {
			_, err := c.Teams.RemoveEscalationPolicy(teamID, ep)
			if err != nil && !isErrCode(err, 404) {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(err)
			}
			return nil
		})
		if retryErr != nil {
			if !isErrCode(retryErr, 404) {
				return nil, fmt.Errorf("%w; Error while trying to dissociate Team %q from Escalation Policy %q", retryErr, teamID, ep)
			} else {
				// Skip Escaltion Policies not found. This happens when a destroy
				// operation is requested and Escalation Policy is destroyed first.
				fmt.Println("SKIP here")
				continue
			}
		}
		epsDissociatedFromTeam = append(epsDissociatedFromTeam, ep)
		fmt.Printf("[DEBUG] EscalationPolicy %s removed from team %s\n", ep, teamID)
	}
	return epsDissociatedFromTeam, nil
}

func associateEPsBackToTeam(c *pagerduty.Client, teamID string, eps []string) error {
	for _, ep := range eps {
		retryErr := resource.Retry(10*time.Second, func() *resource.RetryError {
			_, err := c.Teams.AddEscalationPolicy(teamID, ep)
			if err != nil && !isErrCode(err, 404) {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(err)
			}
			return nil
		})
		if retryErr != nil {
			continue
		}
		fmt.Printf("[DEBUG] EscalationPolicy %s added to team %s\n", ep, teamID)
	}
	return nil
}
