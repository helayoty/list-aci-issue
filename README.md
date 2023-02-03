# List Container Groups Issue

This command utility to help us reproduce the NewListByResourceGroupPager issue with nil instanceView and IPAddress.

1. Clone the repo

```shell
git clone https://github.com/helayoty/list-aci-issue.git

cd list-aci-issue
```

2. Run the command utility

```shell
go run main.go
```

3. Enter the following inputs

> **Note** `SubscriptionID` is mandatory input.

```shell
Please enter the SubscriptionID: #####
#####
```
4. The first run should have no issues. Now, repeat steps 1-3 again and you will start notice the InstanceView will be nil for recreated container groups.
