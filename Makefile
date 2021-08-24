FILE= "terraform-provider-sealedsecret"
local:
	go build -o $(FILE) && mv $(FILE) ~/.terraform.d/plugins/terraform.example.com/local/sealedsecret/0.0.1/linux_amd64/

