---
layout: "mcs"
page_title: "mcs: db_user"
sidebar_current: "docs-db-user"
subcategory: ""
description: |-
  Manages a db user.
---

# mcs\_db\_user

Provides a db user resource. This can be used to create, modify and delete db user.

## Example Usage

```terraform

resource "mcs_db_user" "myuser" {
  name        = "myuser"
  password    = "password"
  
  instance_id = example_db_instance_id
  
  databases   = [example_db_database_name, example_db_other_database_name]
}
```
## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the user. Changing this creates a new user.

* `password` - (Required) The password of the user.

* `host` - IP address of the host that user will be accessible from.

* `instance_id` - (Required) ID of the instance that user is created for. Changing this creates a new user.

* `databases` - List of names of the databases, that user is created for.
