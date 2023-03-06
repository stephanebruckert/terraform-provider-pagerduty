package pagerduty

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/heimweh/go-pagerduty/pagerduty"
)

func TestAccPagerDutyTeamMembership_Basic(t *testing.T) {
	user := fmt.Sprintf("tf-%s", acctest.RandString(5))
	team := fmt.Sprintf("tf-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyTeamMembershipDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyTeamMembershipConfig(user, team),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyTeamMembershipExists("pagerduty_team_membership.foo"),
				),
			},
		},
	})
}

func TestAccPagerDutyTeamMembership_WithRole(t *testing.T) {
	user := fmt.Sprintf("tf-%s", acctest.RandString(5))
	team := fmt.Sprintf("tf-%s", acctest.RandString(5))
	role := "manager"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyTeamMembershipDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyTeamMembershipWithRoleConfig(user, team, role),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyTeamMembershipExists("pagerduty_team_membership.foo"),
				),
			},
		},
	})
}

func TestAccPagerDutyTeamMembership_WithRoleConsistentlyAssigned(t *testing.T) {
	user := fmt.Sprintf("tf-%s", acctest.RandString(5))
	team := fmt.Sprintf("tf-%s", acctest.RandString(5))
	firstRole := "observer"
	secondRole := "responder"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyTeamMembershipDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyTeamMembershipWithRoleConfig(user, team, firstRole),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyTeamMembershipExists("pagerduty_team_membership.foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_team_membership.foo", "role", firstRole),
				),
			},
			{
				Config: testAccCheckPagerDutyTeamMembershipWithRoleConfig(user, team, secondRole),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyTeamMembershipExists("pagerduty_team_membership.foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_team_membership.foo", "role", secondRole),
				),
			},
		},
	})
}

func TestAccPagerDutyTeamMembership_DestroyWithEscalationPolicyDependant(t *testing.T) {
	user := fmt.Sprintf("tf-%s", acctest.RandString(5))
	team := fmt.Sprintf("tf-%s", acctest.RandString(5))
	role := "manager"
	escalationPolicy := fmt.Sprintf("tf-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyTeamMembershipDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyTeamMembershipDestroyWithEscalationPolicyDependant(user, team, role, escalationPolicy),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyTeamMembershipExists("pagerduty_team_membership.foo"),
				),
			},
			{
				Config: testAccCheckPagerDutyTeamMembershipDestroyWithEscalationPolicyDependantUpdated(user, team, role, escalationPolicy),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyTeamMembershipNoExists("pagerduty_team_membership.foo"),
				),
			},
		},
	})
}

func TestAccPagerDutyTeamMembership_DestroyWithUserAndTeam(t *testing.T) {
	user := fmt.Sprintf("tf-%s", acctest.RandString(5))
	user2 := fmt.Sprintf("tf-%s", acctest.RandString(5))

	team := fmt.Sprintf("tf-%s", acctest.RandString(5))
	role := "manager"
	escalationPolicy := fmt.Sprintf("tf-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyTeamMembershipDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyTeamMembershipDestroyWithUserAndTeam(user, team, role, escalationPolicy, user2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyTeamMembershipExists("pagerduty_team_membership.foo"),
					testAccCheckPagerDutyTeamExists("pagerduty_team.foo"),
					testAccCheckPagerDutyUserExists("pagerduty_user.foo"),
				),
			},
			{
				Config: testAccCheckPagerDutyTeamMembershipDestroyWithUserAndTeamUpdated(user, team, role, escalationPolicy, user2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyTeamMembershipNoExists("pagerduty_team_membership.foo"),
					testAccCheckPagerDutyTeamExists("pagerduty_team.foo"),
					testAccCheckPagerDutyUserExists("pagerduty_user.foo"),
				),
			},
		},
	})
}

func TestAccPagerDutyTeamMembership_DestroyWithUsersAndEscalationPolicyDependant(t *testing.T) {
	user := fmt.Sprintf("tf-%s", acctest.RandString(5))
	user2 := fmt.Sprintf("tf-%s", acctest.RandString(5))

	team := fmt.Sprintf("tf-%s", acctest.RandString(5))
	role := "manager"
	escalationPolicy := fmt.Sprintf("tf-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyTeamMembershipDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyTeamMembershipDestroyWithUsersAndEscalationPolicyDependant(user, team, role, escalationPolicy, user2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyTeamMembershipExists("pagerduty_team_membership.foo"),
					testAccCheckPagerDutyUserExists("pagerduty_user.foo"),
				),
			},
			{
				Config: testAccCheckPagerDutyTeamMembershipDestroyWithUsersAndEscalationPolicyDependantUpdated(user, team, role, escalationPolicy, user2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyTeamMembershipNoExists("pagerduty_team_membership.bar"),
					testAccCheckPagerDutyUserNoExists("pagerduty_user.bar"),
					testAccCheckPagerDutyUserExists("pagerduty_user.foo"),
				),
			},
		},
	})
}

func testAccCheckPagerDutyTeamMembershipDestroy(s *terraform.State) error {
	fmt.Println("testAccCheckPagerDutyTeamMembershipDestroy")
	client, _ := testAccProvider.Meta().(*Config).Client()
	for _, r := range s.RootModule().Resources {
		if r.Type != "pagerduty_team_membership" {
			continue
		}

		user, _, err := client.Users.Get(r.Primary.Attributes["user_id"], &pagerduty.GetUserOptions{})
		if err == nil {
			if isTeamMember(user, r.Primary.Attributes["team_id"]) {
				return fmt.Errorf("%s is still a member of: %s", user.ID, r.Primary.Attributes["team_id"])
			}
		}
	}

	return nil
}

func testAccCheckPagerDutyTeamMembershipExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, _ := testAccProvider.Meta().(*Config).Client()
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		userID := rs.Primary.Attributes["user_id"]
		teamID := rs.Primary.Attributes["team_id"]
		role := rs.Primary.Attributes["role"]

		user, _, err := client.Users.Get(userID, &pagerduty.GetUserOptions{})
		if err != nil {
			return err
		}

		if !isTeamMember(user, teamID) {
			return fmt.Errorf("%s is not a member of: %s", userID, teamID)
		}

		resp, _, err := client.Teams.GetMembers(teamID, &pagerduty.GetMembersOptions{})
		if err != nil {
			return err
		}

		for _, member := range resp.Members {
			if member.User.ID == userID {
				if member.Role != role {
					return fmt.Errorf("%s does not have the role: %s in: %s", userID, role, teamID)
				}
			}
		}

		return nil
	}
}

func testAccCheckPagerDutyTeamMembershipNoExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, _ := testAccProvider.Meta().(*Config).Client()
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return nil
		}

		if rs.Primary.ID == "" {
			return nil
		}

		userID := rs.Primary.Attributes["user_id"]
		teamID := rs.Primary.Attributes["team_id"]

		user, _, err := client.Users.Get(userID, &pagerduty.GetUserOptions{})
		if err != nil {
			return err
		}

		if isTeamMember(user, teamID) {
			return fmt.Errorf("%s is still a member of: %s", userID, teamID)
		}

		return nil
	}
}

func testAccCheckPagerDutyTeamMembershipConfig(user, team string) string {
	return fmt.Sprintf(`
resource "pagerduty_user" "foo" {
  name = "%[1]v"
  email = "%[1]v@dazn.com"
}

resource "pagerduty_team" "foo" {
  name        = "%[2]v"
  description = "foo"
}

resource "pagerduty_team_membership" "foo" {
  user_id = pagerduty_user.foo.id
  team_id = pagerduty_team.foo.id
}
`, user, team)
}

func testAccCheckPagerDutyTeamMembershipWithRoleConfig(user, team, role string) string {
	return fmt.Sprintf(`
resource "pagerduty_user" "foo" {
  name = "%[1]v"
  email = "%[1]v@dazn.com"
}

resource "pagerduty_team" "foo" {
  name        = "%[2]v"
  description = "foo"
}

resource "pagerduty_team_membership" "foo" {
  user_id = pagerduty_user.foo.id
  team_id = pagerduty_team.foo.id
  role    = "%[3]v"
}
`, user, team, role)
}

func testAccCheckPagerDutyTeamMembershipDestroyWithEscalationPolicyDependant(user, team, role, escalationPolicy string) string {
	return fmt.Sprintf(`
resource "pagerduty_user" "foo" {
  name = "%[1]v"
  email = "%[1]v@dazn.com"
}

resource "pagerduty_team" "foo" {
  name        = "%[2]v"
  description = "foo"
}

resource "pagerduty_team_membership" "foo" {
  user_id = pagerduty_user.foo.id
  team_id = pagerduty_team.foo.id
  role    = "%[3]v"
}

resource "pagerduty_escalation_policy" "foo" {
  name      = "%[4]v"
  num_loops = 2
  teams     = [pagerduty_team.foo.id]

  rule {
    escalation_delay_in_minutes = 10
    target {
      type = "user_reference"
      id   = pagerduty_user.foo.id
    }
  }
}
`, user, team, role, escalationPolicy)
}

func testAccCheckPagerDutyTeamMembershipDestroyWithEscalationPolicyDependantUpdated(user, team, role, escalationPolicy string) string {
	return fmt.Sprintf(`
resource "pagerduty_user" "foo" {
  name = "%[1]v"
  email = "%[1]v@dazn.com"
}

resource "pagerduty_team" "foo" {
  name        = "%[2]v"
  description = "foo"
}

resource "pagerduty_escalation_policy" "foo" {
  name      = "%[4]s"
  num_loops = 2
  teams     = [pagerduty_team.foo.id]

  rule {
    escalation_delay_in_minutes = 10
    target {
      type = "user_reference"
      id   = pagerduty_user.foo.id
    }
  }
}
`, user, team, role, escalationPolicy)
}

func testAccCheckPagerDutyTeamMembershipDestroyWithUserAndTeam(user, team, role, escalationPolicy, otherUser string) string {
	location := "America/New_York"

	start := timeNowInLoc(location).Add(24 * time.Hour).Round(1 * time.Hour).Format(time.RFC3339)
	rotationVirtualStart := timeNowInLoc(location).Add(24 * time.Hour).Round(1 * time.Hour).Format(time.RFC3339)

	return fmt.Sprintf(`
resource "pagerduty_team" "foo" {
  name        = "%[2]v"
  description = "foo"
}

resource "pagerduty_team_membership" "foo" {
  user_id = pagerduty_user.foo.id
  team_id = pagerduty_team.foo.id
  role    = "%[3]v"
}

resource "pagerduty_user" "foo" {
  name = "%[1]v"
  email = "%[1]vfoo@dazn.com"
}

resource "pagerduty_team_membership" "bar" {
  user_id = pagerduty_user.bar.id
  team_id = pagerduty_team.foo.id
  role    = "%[3]v"
}

resource "pagerduty_user" "bar" {
  name = "%[5]v"
  email = "%[5]vbar@dazn.com"
}

resource "pagerduty_escalation_policy" "foo" {
  name      = "%[4]s"
  num_loops = 2
  teams     = [pagerduty_team.foo.id]

  rule {
    escalation_delay_in_minutes = 10
    target {
      type = "user_reference"
      id   = pagerduty_user.foo.id
    }
  }

  rule {
    escalation_delay_in_minutes = 10
    target {
      type = "user_reference"
      id   = pagerduty_user.bar.id
    }
  }
}

resource "pagerduty_schedule" "foo" {
  name = "%[4]s"

  time_zone   = "America/New_York"
  description = "foo"

  layer {
    name                         = "foo"
    start                        = "%[6]s"
    rotation_virtual_start       = "%[7]s"
    rotation_turn_length_seconds = 86400
    users                        = [pagerduty_user.foo.id, pagerduty_user.bar.id]

    restriction {
      type              = "daily_restriction"
      start_time_of_day = "08:00:00"
      duration_seconds  = 32101
    }
  }
}
`, user, team, role, escalationPolicy, otherUser, start, rotationVirtualStart)
}

func testAccCheckPagerDutyTeamMembershipDestroyWithUserAndTeamUpdated(user, team, role, escalationPolicy, otherUser string) string {
	location := "America/New_York"

	start := timeNowInLoc(location).Add(24 * time.Hour).Round(1 * time.Hour).Format(time.RFC3339)
	rotationVirtualStart := timeNowInLoc(location).Add(24 * time.Hour).Round(1 * time.Hour).Format(time.RFC3339)

	return fmt.Sprintf(`
resource "pagerduty_team" "foo" {
  name        = "%[2]v"
  description = "foo"
}

resource "pagerduty_user" "foo" {
  name = "%[1]v"
  email = "%[1]vfoo@dazn.com"
}

resource "pagerduty_escalation_policy" "foo" {
  name      = "%[4]s"
  num_loops = 2
  teams     = [pagerduty_team.foo.id]

  rule {
    escalation_delay_in_minutes = 10
    target {
      type = "user_reference"
      id   = pagerduty_user.foo.id
    }
  }
}

resource "pagerduty_schedule" "foo" {
  name = "%[4]s"

  time_zone   = "America/New_York"
  description = "foo"

  layer {
    name                         = "foo"
    start                        = "%[6]s"
    rotation_virtual_start       = "%[7]s"
    rotation_turn_length_seconds = 86400
    users                        = [pagerduty_user.foo.id]

    restriction {
      type              = "daily_restriction"
      start_time_of_day = "08:00:00"
      duration_seconds  = 32101
    }
  }
}
`, user, team, role, escalationPolicy, otherUser, start, rotationVirtualStart)
}

func testAccCheckPagerDutyTeamMembershipDestroyWithUsersAndEscalationPolicyDependant(user, team, role, escalationPolicy, otherUser string) string {
	return fmt.Sprintf(`

resource "pagerduty_user" "foo" {
  name = "%[1]v"
  email = "%[1]v@dazn.com"
}

resource "pagerduty_team" "foo" {
  name        = "%[2]v"
  description = "foo"
}

resource "pagerduty_team_membership" "foo" {
  user_id = pagerduty_user.foo.id
  team_id = pagerduty_team.foo.id
  role    = "%[3]v"
}

resource "pagerduty_escalation_policy" "foo" {
  name      = "%[5]v"
  num_loops = 2
  teams     = [pagerduty_team.foo.id]

  rule {
    escalation_delay_in_minutes = 10
    target {
      type = "user_reference"
      id   = pagerduty_user.foo.id
    }
  }
}
`, user, team, role, escalationPolicy, otherUser)
}

func testAccCheckPagerDutyTeamMembershipDestroyWithUsersAndEscalationPolicyDependantUpdated(user, team, role, escalationPolicy, otherUser string) string {
	return fmt.Sprintf(`
resource "pagerduty_team" "foo" {
  name        = "%[2]v"
  description = "foo"
}
`, user, team, role, escalationPolicy, otherUser)
}

func isTeamMember(user *pagerduty.User, teamID string) bool {
	var found bool

	for _, team := range user.Teams {
		if teamID == team.ID {
			found = true
		}
	}

	return found
}
