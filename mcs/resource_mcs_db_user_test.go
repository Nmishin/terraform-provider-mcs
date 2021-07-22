package mcs

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gophercloud/gophercloud/openstack/db/v1/users"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccDatabaseUser_basic(t *testing.T) {
	var user users.User
	var instance instanceResp

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckDatabase(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDatabaseUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseUserBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseInstanceExists(
						"mcs_db_instance.basic", &instance),
					testAccCheckDatabaseUserExists(
						"mcs_db_user.basic", &instance, &user),
					resource.TestCheckResourceAttrPtr(
						"mcs_db_user.basic", "name", &user.Name),
				),
			},
		},
	})
}

func TestAccDatabaseUser_update_and_delete(t *testing.T) {
	var user users.User
	var instance instanceResp

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckDatabase(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDatabaseUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseUserBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseInstanceExists(
						"mcs_db_instance.basic", &instance),
					testAccCheckDatabaseUserExists(
						"mcs_db_user.basic", &instance, &user),
					resource.TestCheckResourceAttrPtr(
						"mcs_db_user.basic", "name", &user.Name),
				),
			},
			{
				Config: testAccDatabaseUserAddDatabase,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseUserExists(
						"mcs_db_user.basic", &instance, &user),
					testAccCheckDatabaseUserDatabaseCount(2, &user),
				),
			},
			{
				Config: testAccDatabaseUserBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseUserExists(
						"mcs_db_user.basic", &instance, &user),
					testAccCheckDatabaseUserDatabaseCount(1, &user),
				),
			},
		},
	})
}

func testAccCheckDatabaseUserExists(n string, instance *instanceResp, user *users.User) resource.TestCheckFunc {

	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		parts := strings.SplitN(rs.Primary.ID, "/", 2)
		if len(parts) != 2 {
			return fmt.Errorf("Malformed user name: %s", rs.Primary.ID)
		}

		config := testAccProvider.Meta().(Config)
		DatabaseClient, err := config.DatabaseV1Client(OSRegionName)
		if err != nil {
			return fmt.Errorf("Error creating cloud database client: %s", err)
		}

		pages, err := userList(DatabaseClient, instance.ID).AllPages()
		if err != nil {
			return fmt.Errorf("Unable to retrieve users: %s", err)
		}

		allUsers, err := users.ExtractUsers(pages)
		if err != nil {
			return fmt.Errorf("Unable to extract users: %s", err)
		}

		for _, u := range allUsers {
			if u.Name == parts[1] {
				*user = u
				return nil
			}
		}

		return fmt.Errorf("User %s does not exist", n)
	}
}

func testAccCheckDatabaseUserDatabaseCount(n int, user *users.User) resource.TestCheckFunc {

	return func(s *terraform.State) error {
		if len(user.Databases) != n {
			return fmt.Errorf("Wrong number of databases assigned to user: %s", user.Name)
		}
		return nil
	}
}

func testAccCheckDatabaseUserDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(Config)

	DatabaseClient, err := config.DatabaseV1Client(OSRegionName)
	if err != nil {
		return fmt.Errorf("Error creating cloud database client: %s", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "mcs_db_user" {
			continue
		}

		parts := strings.SplitN(rs.Primary.ID, "/", 2)
		if len(parts) != 2 {
			return fmt.Errorf("Malformed username: %s", rs.Primary.ID)
		}

		pages, err := userList(DatabaseClient, parts[0]).AllPages()
		if err != nil {
			return nil
		}

		allUsers, err := users.ExtractUsers(pages)
		if err != nil {
			return fmt.Errorf("Unable to extract users: %s", err)
		}

		var exists bool
		for _, v := range allUsers {
			if v.Name == parts[1] {
				exists = true
			}
		}

		if exists {
			return fmt.Errorf("User still exists")
		}
	}

	return nil
}

var testAccDatabaseUserBasic = fmt.Sprintf(`
resource "mcs_db_instance" "basic" {
  name = "basic"
  flavor_id = "%s"
  size = 10
  volume_type = "ms1"

  datastore {
    version = "%s"
    type    = "%s"
  }

  network {
    uuid = "%s"
  }
}

resource "mcs_db_database" "testdb1" {
  name = "testdb1"
  instance_id = "${mcs_db_instance.basic.id}"
}
  
resource "mcs_db_database" "testdb2" {
  name = "testdb2"
  instance_id = "${mcs_db_instance.basic.id}"
}

resource "mcs_db_user" "basic" {
  name        = "basic"
  instance_id = "${mcs_db_instance.basic.id}"
  password    = "password"
  databases = [
	"${mcs_db_database.testdb1.name}"
  ]
}
`, OSFlavorID, OSDBDatastoreVersion, OSDBDatastoreType, OSNetworkID)

var testAccDatabaseUserAddDatabase = fmt.Sprintf(`
resource "mcs_db_instance" "basic" {
  name = "basic"
  flavor_id = "%s"
  size = 10
  volume_type = "ms1"

  datastore {
    version = "%s"
    type    = "%s"
  }

  network {
    uuid = "%s"
  }
}

resource "mcs_db_database" "testdb1" {
	name = "testdb1"
	instance_id = "${mcs_db_instance.basic.id}"
}
  
resource "mcs_db_database" "testdb2" {
	name = "testdb2"
	instance_id = "${mcs_db_instance.basic.id}"
}

resource "mcs_db_user" "basic" {
  name        = "basic"
  instance_id = "${mcs_db_instance.basic.id}"
  password    = "password"
  databases = [
	  "${mcs_db_database.testdb2.name}",
	  "${mcs_db_database.testdb1.name}"
  ]
}
`, OSFlavorID, OSDBDatastoreVersion, OSDBDatastoreType, OSNetworkID)
