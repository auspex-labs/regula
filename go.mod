module github.com/fugue/regula

go 1.16

require (
	github.com/alexeyco/simpletable v1.0.0
	github.com/fatih/color v1.9.0
	github.com/golang/mock v1.5.0
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-version v1.3.0 // indirect
	github.com/hashicorp/hcl/v2 v2.10.0
	github.com/hashicorp/terraform v0.15.1
	github.com/hashicorp/terraform-provider-google v1.20.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/libgit2/git2go/v31 v31.4.14
	github.com/mattn/go-colorable v0.1.7 // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/open-policy-agent/opa v0.26.0
	github.com/sergi/go-diff v1.1.0 // indirect
	github.com/spf13/cobra v1.1.3
	github.com/stretchr/testify v1.6.1
	github.com/terraform-providers/terraform-provider-aws v1.60.0
	github.com/terraform-providers/terraform-provider-google v1.20.0
	github.com/thediveo/enumflag v0.10.1
	github.com/zclconf/go-cty v1.8.2
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2
	golang.org/x/mod v0.4.1 // indirect
	golang.org/x/oauth2 v0.0.0-20210220000619-9bb904979d93 // indirect
	golang.org/x/sys v0.0.0-20210305230114-8fe3ee5dd75b // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.26.0 // indirect
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c
	tf_resource_schemas v0.0.0-00010101000000-000000000000
)

replace (
	github.com/terraform-providers/terraform-provider-aws => ./providers/terraform-provider-aws
	github.com/terraform-providers/terraform-provider-google => ./providers/terraform-provider-google
	tf_resource_schemas => ./pkg/tf_resource_schemas/
)
