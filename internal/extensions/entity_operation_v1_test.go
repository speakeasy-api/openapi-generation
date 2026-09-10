package extensions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestParseEntityOperationV1(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		expected *EntityOperationV1
		wantErr  bool
	}{
		{
			name: "scalar node - single operation type",
			yaml: `User#read`,
			expected: &EntityOperationV1{
				TerraformActions: []EntityOperationV1Config{},
				TerraformDataResources: []EntityOperationV1Config{
					{
						Entity:         "User",
						OperationTypes: []string{"read"},
						Order:          nil,
					},
				},
				TerraformEphemeralResources: []EntityOperationV1Config{},
				TerraformManagedResources: []EntityOperationV1Config{
					{
						Entity:         "User",
						OperationTypes: []string{"read"},
						Order:          nil,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "scalar node - multiple operation types",
			yaml: `User#create,update`,
			expected: &EntityOperationV1{
				TerraformActions:            []EntityOperationV1Config{},
				TerraformDataResources:      []EntityOperationV1Config{},
				TerraformEphemeralResources: []EntityOperationV1Config{},
				TerraformManagedResources: []EntityOperationV1Config{
					{
						Entity:         "User",
						OperationTypes: []string{"create", "update"},
						Order:          nil,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "scalar node - with order",
			yaml: `User#read#1`,
			expected: &EntityOperationV1{
				TerraformActions: []EntityOperationV1Config{},
				TerraformDataResources: []EntityOperationV1Config{
					{
						Entity:         "User",
						OperationTypes: []string{"read"},
						Order:          ptr(1),
					},
				},
				TerraformEphemeralResources: []EntityOperationV1Config{},
				TerraformManagedResources: []EntityOperationV1Config{
					{
						Entity:         "User",
						OperationTypes: []string{"read"},
						Order:          ptr(1),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "scalar node - legacy get operation type",
			yaml: `User#get`,
			expected: &EntityOperationV1{
				TerraformActions: []EntityOperationV1Config{},
				TerraformDataResources: []EntityOperationV1Config{
					{
						Entity:         "User",
						OperationTypes: []string{"read"},
						Order:          nil,
					},
				},
				TerraformEphemeralResources: []EntityOperationV1Config{},
				TerraformManagedResources: []EntityOperationV1Config{
					{
						Entity:         "User",
						OperationTypes: []string{"read"},
						Order:          nil,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "sequence node - multiple entities",
			yaml: `
- User#read
- Role#delete
- Permission#read#2`,
			expected: &EntityOperationV1{
				TerraformActions: []EntityOperationV1Config{},
				TerraformDataResources: []EntityOperationV1Config{
					{
						Entity:         "User",
						OperationTypes: []string{"read"},
						Order:          nil,
					},
					{
						Entity:         "Permission",
						OperationTypes: []string{"read"},
						Order:          ptr(2),
					},
				},
				TerraformEphemeralResources: []EntityOperationV1Config{},
				TerraformManagedResources: []EntityOperationV1Config{
					{
						Entity:         "User",
						OperationTypes: []string{"read"},
						Order:          nil,
					},
					{
						Entity:         "Role",
						OperationTypes: []string{"delete"},
						Order:          nil,
					},
					{
						Entity:         "Permission",
						OperationTypes: []string{"read"},
						Order:          ptr(2),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "scalar node - only managed resource operations",
			yaml: `User#create,update,delete`,
			expected: &EntityOperationV1{
				TerraformActions:            []EntityOperationV1Config{},
				TerraformDataResources:      []EntityOperationV1Config{},
				TerraformEphemeralResources: []EntityOperationV1Config{},
				TerraformManagedResources: []EntityOperationV1Config{
					{
						Entity:         "User",
						OperationTypes: []string{"create", "update", "delete"},
						Order:          nil,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "mapping node - terraform-datasource only",
			yaml: `
terraform-datasource: User#read`,
			expected: &EntityOperationV1{
				TerraformActions: []EntityOperationV1Config{},
				TerraformDataResources: []EntityOperationV1Config{
					{
						Entity:         "User",
						OperationTypes: []string{"read"},
						Order:          nil,
					},
				},
				TerraformEphemeralResources: []EntityOperationV1Config{},
				TerraformManagedResources:   []EntityOperationV1Config{},
			},
			wantErr: false,
		},
		{
			name: "mapping node - terraform-resource only",
			yaml: `
terraform-resource: User#create,update`,
			expected: &EntityOperationV1{
				TerraformActions:            []EntityOperationV1Config{},
				TerraformDataResources:      []EntityOperationV1Config{},
				TerraformEphemeralResources: []EntityOperationV1Config{},
				TerraformManagedResources: []EntityOperationV1Config{
					{
						Entity:         "User",
						OperationTypes: []string{"create", "update"},
						Order:          nil,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "mapping node - both terraform-datasource and terraform-resource",
			yaml: `
terraform-datasource: User#read
terraform-resource: Other#read`,
			expected: &EntityOperationV1{
				TerraformActions: []EntityOperationV1Config{},
				TerraformDataResources: []EntityOperationV1Config{
					{
						Entity:         "User",
						OperationTypes: []string{"read"},
						Order:          nil,
					},
				},
				TerraformEphemeralResources: []EntityOperationV1Config{},
				TerraformManagedResources: []EntityOperationV1Config{
					{
						Entity:         "Other",
						OperationTypes: []string{"read"},
						Order:          nil,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "mapping node - with order",
			yaml: `
terraform-datasource: User#read#1
terraform-resource: User#read#2`,
			expected: &EntityOperationV1{
				TerraformActions: []EntityOperationV1Config{},
				TerraformDataResources: []EntityOperationV1Config{
					{
						Entity:         "User",
						OperationTypes: []string{"read"},
						Order:          ptr(1),
					},
				},
				TerraformEphemeralResources: []EntityOperationV1Config{},
				TerraformManagedResources: []EntityOperationV1Config{
					{
						Entity:         "User",
						OperationTypes: []string{"read"},
						Order:          ptr(2),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "mapping node - terraform-datasource with sequence",
			yaml: `
terraform-datasource:
  - User#read
  - Role#read#1`,
			expected: &EntityOperationV1{
				TerraformActions: []EntityOperationV1Config{},
				TerraformDataResources: []EntityOperationV1Config{
					{
						Entity:         "User",
						OperationTypes: []string{"read"},
						Order:          nil,
					},
					{
						Entity:         "Role",
						OperationTypes: []string{"read"},
						Order:          ptr(1),
					},
				},
				TerraformEphemeralResources: []EntityOperationV1Config{},
				TerraformManagedResources:   []EntityOperationV1Config{},
			},
			wantErr: false,
		},
		{
			name: "mapping node - terraform-resource with sequence",
			yaml: `
terraform-resource:
  - User#create,update
  - Role#read#2`,
			expected: &EntityOperationV1{
				TerraformActions:            []EntityOperationV1Config{},
				TerraformDataResources:      []EntityOperationV1Config{},
				TerraformEphemeralResources: []EntityOperationV1Config{},
				TerraformManagedResources: []EntityOperationV1Config{
					{
						Entity:         "User",
						OperationTypes: []string{"create", "update"},
						Order:          nil,
					},
					{
						Entity:         "Role",
						OperationTypes: []string{"read"},
						Order:          ptr(2),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "mapping node - both with sequences",
			yaml: `
terraform-datasource:
  - User#read
  - Permission#read#1
terraform-resource:
  - Role#read
  - Status#read#2`,
			expected: &EntityOperationV1{
				TerraformActions: []EntityOperationV1Config{},
				TerraformDataResources: []EntityOperationV1Config{
					{
						Entity:         "User",
						OperationTypes: []string{"read"},
						Order:          nil,
					},
					{
						Entity:         "Permission",
						OperationTypes: []string{"read"},
						Order:          ptr(1),
					},
				},
				TerraformEphemeralResources: []EntityOperationV1Config{},
				TerraformManagedResources: []EntityOperationV1Config{
					{
						Entity:         "Role",
						OperationTypes: []string{"read"},
						Order:          nil,
					},
					{
						Entity:         "Status",
						OperationTypes: []string{"read"},
						Order:          ptr(2),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "mapping node - terraform-resource with separate operations per entity",
			yaml: `
terraform-resource:
  - User#create
  - User#update`,
			expected: &EntityOperationV1{
				TerraformActions:            []EntityOperationV1Config{},
				TerraformDataResources:      []EntityOperationV1Config{},
				TerraformEphemeralResources: []EntityOperationV1Config{},
				TerraformManagedResources: []EntityOperationV1Config{
					{
						Entity:         "User",
						OperationTypes: []string{"create"},
						Order:          nil,
					},
					{
						Entity:         "User",
						OperationTypes: []string{"update"},
						Order:          nil,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "mapping node - with null value",
			yaml: `
terraform-resource: User#create
terraform-datasource: null`,
			expected: &EntityOperationV1{
				TerraformActions:            []EntityOperationV1Config{},
				TerraformDataResources:      []EntityOperationV1Config{},
				TerraformEphemeralResources: []EntityOperationV1Config{},
				TerraformManagedResources: []EntityOperationV1Config{
					{
						Entity:         "User",
						OperationTypes: []string{"create"},
						Order:          nil,
					},
				},
			},
			wantErr: false,
		},
		{
			name:     "invalid scalar node - malformed format",
			yaml:     `User`,
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "invalid scalar node - invalid operation type",
			yaml:     `User#invalid`,
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "invalid scalar node - invalid order",
			yaml:     `User#read#invalid`,
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "invalid scalar node - zero order",
			yaml:     `User#read#0`,
			expected: nil,
			wantErr:  true,
		},
		{
			name: "invalid sequence node - malformed item",
			yaml: `
- User#read
- Invalid`,
			expected: nil,
			wantErr:  true,
		},
		{
			name: "invalid mapping node - malformed terraform-datasource",
			yaml: `
terraform-datasource: Invalid`,
			expected: nil,
			wantErr:  true,
		},
		{
			name: "invalid mapping node - malformed terraform-resource",
			yaml: `
terraform-resource: Invalid`,
			expected: nil,
			wantErr:  true,
		},
		{
			name: "empty mapping node",
			yaml: `{}`,
			expected: &EntityOperationV1{
				TerraformActions:            []EntityOperationV1Config{},
				TerraformDataResources:      []EntityOperationV1Config{},
				TerraformEphemeralResources: []EntityOperationV1Config{},
				TerraformManagedResources:   []EntityOperationV1Config{},
			},
			wantErr: false,
		},
		{
			name: "sequence node - empty array",
			yaml: `[]`,
			expected: &EntityOperationV1{
				TerraformActions:            []EntityOperationV1Config{},
				TerraformDataResources:      []EntityOperationV1Config{},
				TerraformEphemeralResources: []EntityOperationV1Config{},
				TerraformManagedResources:   []EntityOperationV1Config{},
			},
			wantErr: false,
		},
		{
			name: "sequence node - mapping form with options",
			yaml: `
- entityOperation: User#read
  options:
    polling:
      name: UserPolling`,
			expected: &EntityOperationV1{
				TerraformActions: []EntityOperationV1Config{},
				TerraformDataResources: []EntityOperationV1Config{
					{
						Entity:         "User",
						OperationTypes: []string{"read"},
						Order:          nil,
						Options: &EntityOperationV1Options{
							Polling: &EntityOperationV1Polling{
								Name: "UserPolling",
							},
						},
					},
				},
				TerraformEphemeralResources: []EntityOperationV1Config{},
				TerraformManagedResources: []EntityOperationV1Config{
					{
						Entity:         "User",
						OperationTypes: []string{"read"},
						Order:          nil,
						Options: &EntityOperationV1Options{
							Polling: &EntityOperationV1Polling{
								Name: "UserPolling",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "sequence node - mixed scalar and mapping forms",
			yaml: `
- User#read
- entityOperation: Role#create,update
  options:
    polling:
      name: RolePolling
- Permission#read#2`,
			expected: &EntityOperationV1{
				TerraformActions: []EntityOperationV1Config{},
				TerraformDataResources: []EntityOperationV1Config{
					{
						Entity:         "User",
						OperationTypes: []string{"read"},
						Order:          nil,
						Options:        nil,
					},
					{
						Entity:         "Permission",
						OperationTypes: []string{"read"},
						Order:          ptr(2),
						Options:        nil,
					},
				},
				TerraformEphemeralResources: []EntityOperationV1Config{},
				TerraformManagedResources: []EntityOperationV1Config{
					{
						Entity:         "User",
						OperationTypes: []string{"read"},
						Order:          nil,
						Options:        nil,
					},
					{
						Entity:         "Role",
						OperationTypes: []string{"create", "update"},
						Order:          nil,
						Options: &EntityOperationV1Options{
							Polling: &EntityOperationV1Polling{
								Name: "RolePolling",
							},
						},
					},
					{
						Entity:         "Permission",
						OperationTypes: []string{"read"},
						Order:          ptr(2),
						Options:        nil,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "sequence node - mapping form with order",
			yaml: `
- entityOperation: User#read#1
  options:
    polling:
      name: UserPolling`,
			expected: &EntityOperationV1{
				TerraformActions: []EntityOperationV1Config{},
				TerraformDataResources: []EntityOperationV1Config{
					{
						Entity:         "User",
						OperationTypes: []string{"read"},
						Order:          ptr(1),
						Options: &EntityOperationV1Options{
							Polling: &EntityOperationV1Polling{
								Name: "UserPolling",
							},
						},
					},
				},
				TerraformEphemeralResources: []EntityOperationV1Config{},
				TerraformManagedResources: []EntityOperationV1Config{
					{
						Entity:         "User",
						OperationTypes: []string{"read"},
						Order:          ptr(1),
						Options: &EntityOperationV1Options{
							Polling: &EntityOperationV1Polling{
								Name: "UserPolling",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "mapping node - terraform-resource with mapping form in sequence",
			yaml: `
terraform-resource:
  - User#create
  - entityOperation: Role#update
    options:
      polling:
        name: RolePolling`,
			expected: &EntityOperationV1{
				TerraformActions:            []EntityOperationV1Config{},
				TerraformDataResources:      []EntityOperationV1Config{},
				TerraformEphemeralResources: []EntityOperationV1Config{},
				TerraformManagedResources: []EntityOperationV1Config{
					{
						Entity:         "User",
						OperationTypes: []string{"create"},
						Order:          nil,
						Options:        nil,
					},
					{
						Entity:         "Role",
						OperationTypes: []string{"update"},
						Order:          nil,
						Options: &EntityOperationV1Options{
							Polling: &EntityOperationV1Polling{
								Name: "RolePolling",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid sequence node - mapping without entityOperation",
			yaml: `
- options:
    polling:
      name: UserPolling`,
			expected: nil,
			wantErr:  true,
		},
		{
			name: "invalid sequence node - mapping with invalid entityOperation",
			yaml: `
- entityOperation: InvalidFormat
  options:
    polling:
      name: UserPolling`,
			expected: nil,
			wantErr:  true,
		},
		{
			name: "invalid sequence node - mapping with missing polling name",
			yaml: `
- entityOperation: User#read
  options:
    polling: {}`,
			expected: nil,
			wantErr:  true,
		},
		{
			name: "sequence node - mapping form with all polling options",
			yaml: `
- entityOperation: User#read
  options:
    polling:
      name: UserPolling
      delaySeconds: 5
      intervalSeconds: 10
      limitCount: 20`,
			expected: &EntityOperationV1{
				TerraformActions: []EntityOperationV1Config{},
				TerraformDataResources: []EntityOperationV1Config{
					{
						Entity:         "User",
						OperationTypes: []string{"read"},
						Order:          nil,
						Options: &EntityOperationV1Options{
							Polling: &EntityOperationV1Polling{
								Name:            "UserPolling",
								DelaySeconds:    ptr(5),
								IntervalSeconds: ptr(10),
								LimitCount:      ptr(20),
							},
						},
					},
				},
				TerraformEphemeralResources: []EntityOperationV1Config{},
				TerraformManagedResources: []EntityOperationV1Config{
					{
						Entity:         "User",
						OperationTypes: []string{"read"},
						Order:          nil,
						Options: &EntityOperationV1Options{
							Polling: &EntityOperationV1Polling{
								Name:            "UserPolling",
								DelaySeconds:    ptr(5),
								IntervalSeconds: ptr(10),
								LimitCount:      ptr(20),
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "sequence node - mapping form with partial polling options",
			yaml: `
- entityOperation: Role#create
  options:
    polling:
      name: RolePolling
      intervalSeconds: 15`,
			expected: &EntityOperationV1{
				TerraformActions:            []EntityOperationV1Config{},
				TerraformDataResources:      []EntityOperationV1Config{},
				TerraformEphemeralResources: []EntityOperationV1Config{},
				TerraformManagedResources: []EntityOperationV1Config{
					{
						Entity:         "Role",
						OperationTypes: []string{"create"},
						Order:          nil,
						Options: &EntityOperationV1Options{
							Polling: &EntityOperationV1Polling{
								Name:            "RolePolling",
								IntervalSeconds: ptr(15),
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "sequence node - mapping form with zero delaySeconds",
			yaml: `
- entityOperation: User#read
  options:
    polling:
      name: UserPolling
      delaySeconds: 0`,
			expected: &EntityOperationV1{
				TerraformActions: []EntityOperationV1Config{},
				TerraformDataResources: []EntityOperationV1Config{
					{
						Entity:         "User",
						OperationTypes: []string{"read"},
						Order:          nil,
						Options: &EntityOperationV1Options{
							Polling: &EntityOperationV1Polling{
								Name:         "UserPolling",
								DelaySeconds: ptr(0),
							},
						},
					},
				},
				TerraformEphemeralResources: []EntityOperationV1Config{},
				TerraformManagedResources: []EntityOperationV1Config{
					{
						Entity:         "User",
						OperationTypes: []string{"read"},
						Order:          nil,
						Options: &EntityOperationV1Options{
							Polling: &EntityOperationV1Polling{
								Name:         "UserPolling",
								DelaySeconds: ptr(0),
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "mapping node - terraform-resource with all polling options",
			yaml: `
terraform-resource:
  - entityOperation: User#create,update,delete
    options:
      polling:
        name: UserPolling
        delaySeconds: 2
        intervalSeconds: 5
        limitCount: 100`,
			expected: &EntityOperationV1{
				TerraformActions:            []EntityOperationV1Config{},
				TerraformDataResources:      []EntityOperationV1Config{},
				TerraformEphemeralResources: []EntityOperationV1Config{},
				TerraformManagedResources: []EntityOperationV1Config{
					{
						Entity:         "User",
						OperationTypes: []string{"create", "update", "delete"},
						Order:          nil,
						Options: &EntityOperationV1Options{
							Polling: &EntityOperationV1Polling{
								Name:            "UserPolling",
								DelaySeconds:    ptr(2),
								IntervalSeconds: ptr(5),
								LimitCount:      ptr(100),
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid sequence node - negative delaySeconds",
			yaml: `
- entityOperation: User#read
  options:
    polling:
      name: UserPolling
      delaySeconds: -1`,
			expected: nil,
			wantErr:  true,
		},
		{
			name: "invalid sequence node - zero intervalSeconds",
			yaml: `
- entityOperation: User#read
  options:
    polling:
      name: UserPolling
      intervalSeconds: 0`,
			expected: nil,
			wantErr:  true,
		},
		{
			name: "invalid sequence node - negative intervalSeconds",
			yaml: `
- entityOperation: User#read
  options:
    polling:
      name: UserPolling
      intervalSeconds: -5`,
			expected: nil,
			wantErr:  true,
		},
		{
			name: "invalid sequence node - zero limitCount",
			yaml: `
- entityOperation: User#read
  options:
    polling:
      name: UserPolling
      limitCount: 0`,
			expected: nil,
			wantErr:  true,
		},
		{
			name: "invalid sequence node - negative limitCount",
			yaml: `
- entityOperation: User#read
  options:
    polling:
      name: UserPolling
      limitCount: -10`,
			expected: nil,
			wantErr:  true,
		},
		{
			name: "invalid sequence node - non-integer delaySeconds",
			yaml: `
- entityOperation: User#read
  options:
    polling:
      name: UserPolling
      delaySeconds: invalid`,
			expected: nil,
			wantErr:  true,
		},
		{
			name: "invalid sequence node - non-integer intervalSeconds",
			yaml: `
- entityOperation: User#read
  options:
    polling:
      name: UserPolling
      intervalSeconds: 3.5`,
			expected: nil,
			wantErr:  true,
		},
		{
			name: "invalid sequence node - non-integer limitCount",
			yaml: `
- entityOperation: User#read
  options:
    polling:
      name: UserPolling
      limitCount: abc`,
			expected: nil,
			wantErr:  true,
		},
		{
			name: "scalar node - ephemeral resource open operation",
			yaml: `Token#open`,
			expected: &EntityOperationV1{
				TerraformActions:       []EntityOperationV1Config{},
				TerraformDataResources: []EntityOperationV1Config{},
				TerraformEphemeralResources: []EntityOperationV1Config{
					{
						Entity:         "Token",
						OperationTypes: []string{"open"},
						Order:          nil,
					},
				},
				TerraformManagedResources: []EntityOperationV1Config{},
			},
			wantErr: false,
		},
		{
			name: "scalar node - ephemeral resource with order",
			yaml: `Token#open#1`,
			expected: &EntityOperationV1{
				TerraformActions:       []EntityOperationV1Config{},
				TerraformDataResources: []EntityOperationV1Config{},
				TerraformEphemeralResources: []EntityOperationV1Config{
					{
						Entity:         "Token",
						OperationTypes: []string{"open"},
						Order:          ptr(1),
					},
				},
				TerraformManagedResources: []EntityOperationV1Config{},
			},
			wantErr: false,
		},
		{
			name: "sequence node - multiple ephemeral resources",
			yaml: `
- Token#open
- Secret#open#1`,
			expected: &EntityOperationV1{
				TerraformActions:       []EntityOperationV1Config{},
				TerraformDataResources: []EntityOperationV1Config{},
				TerraformEphemeralResources: []EntityOperationV1Config{
					{
						Entity:         "Token",
						OperationTypes: []string{"open"},
						Order:          nil,
					},
					{
						Entity:         "Secret",
						OperationTypes: []string{"open"},
						Order:          ptr(1),
					},
				},
				TerraformManagedResources: []EntityOperationV1Config{},
			},
			wantErr: false,
		},
		{
			name: "mapping node - terraform-ephemeral-resource only",
			yaml: `
terraform-ephemeral-resource: Token#open`,
			expected: &EntityOperationV1{
				TerraformActions:       []EntityOperationV1Config{},
				TerraformDataResources: []EntityOperationV1Config{},
				TerraformEphemeralResources: []EntityOperationV1Config{
					{
						Entity:         "Token",
						OperationTypes: []string{"open"},
						Order:          nil,
					},
				},
				TerraformManagedResources: []EntityOperationV1Config{},
			},
			wantErr: false,
		},
		{
			name: "mapping node - terraform-ephemeral-resource with sequence",
			yaml: `
terraform-ephemeral-resource:
  - Token#open
  - Secret#open#2`,
			expected: &EntityOperationV1{
				TerraformActions:       []EntityOperationV1Config{},
				TerraformDataResources: []EntityOperationV1Config{},
				TerraformEphemeralResources: []EntityOperationV1Config{
					{
						Entity:         "Token",
						OperationTypes: []string{"open"},
						Order:          nil,
					},
					{
						Entity:         "Secret",
						OperationTypes: []string{"open"},
						Order:          ptr(2),
					},
				},
				TerraformManagedResources: []EntityOperationV1Config{},
			},
			wantErr: false,
		},
		{
			name: "mapping node - all four terraform types",
			yaml: `
terraform-action: Task#invoke
terraform-datasource: User#read
terraform-ephemeral-resource: Token#open
terraform-resource: Role#create`,
			expected: &EntityOperationV1{
				TerraformActions: []EntityOperationV1Config{
					{
						Entity:         "Task",
						OperationTypes: []string{"invoke"},
						Order:          nil,
					},
				},
				TerraformDataResources: []EntityOperationV1Config{
					{
						Entity:         "User",
						OperationTypes: []string{"read"},
						Order:          nil,
					},
				},
				TerraformEphemeralResources: []EntityOperationV1Config{
					{
						Entity:         "Token",
						OperationTypes: []string{"open"},
						Order:          nil,
					},
				},
				TerraformManagedResources: []EntityOperationV1Config{
					{
						Entity:         "Role",
						OperationTypes: []string{"create"},
						Order:          nil,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "mapping node - all four with sequences",
			yaml: `
terraform-action:
  - Task#invoke
  - Cleanup#invoke#2
terraform-datasource:
  - User#read
  - Permission#read#1
terraform-ephemeral-resource:
  - Token#open
  - Secret#open#2
terraform-resource:
  - Role#create
  - Status#update#3`,
			expected: &EntityOperationV1{
				TerraformActions: []EntityOperationV1Config{
					{
						Entity:         "Task",
						OperationTypes: []string{"invoke"},
						Order:          nil,
					},
					{
						Entity:         "Cleanup",
						OperationTypes: []string{"invoke"},
						Order:          ptr(2),
					},
				},
				TerraformDataResources: []EntityOperationV1Config{
					{
						Entity:         "User",
						OperationTypes: []string{"read"},
						Order:          nil,
					},
					{
						Entity:         "Permission",
						OperationTypes: []string{"read"},
						Order:          ptr(1),
					},
				},
				TerraformEphemeralResources: []EntityOperationV1Config{
					{
						Entity:         "Token",
						OperationTypes: []string{"open"},
						Order:          nil,
					},
					{
						Entity:         "Secret",
						OperationTypes: []string{"open"},
						Order:          ptr(2),
					},
				},
				TerraformManagedResources: []EntityOperationV1Config{
					{
						Entity:         "Role",
						OperationTypes: []string{"create"},
						Order:          nil,
					},
					{
						Entity:         "Status",
						OperationTypes: []string{"update"},
						Order:          ptr(3),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "sequence node - ephemeral resource with mapping form and options",
			yaml: `
- entityOperation: Token#open
  options:
    polling:
      name: TokenPolling
      delaySeconds: 1
      intervalSeconds: 3
      limitCount: 10`,
			expected: &EntityOperationV1{
				TerraformActions:       []EntityOperationV1Config{},
				TerraformDataResources: []EntityOperationV1Config{},
				TerraformEphemeralResources: []EntityOperationV1Config{
					{
						Entity:         "Token",
						OperationTypes: []string{"open"},
						Order:          nil,
						Options: &EntityOperationV1Options{
							Polling: &EntityOperationV1Polling{
								Name:            "TokenPolling",
								DelaySeconds:    ptr(1),
								IntervalSeconds: ptr(3),
								LimitCount:      ptr(10),
							},
						},
					},
				},
				TerraformManagedResources: []EntityOperationV1Config{},
			},
			wantErr: false,
		},
		{
			name: "mapping node - terraform-ephemeral-resource with mapping form in sequence",
			yaml: `
terraform-ephemeral-resource:
  - Token#open
  - entityOperation: Secret#open#1
    options:
      polling:
        name: SecretPolling`,
			expected: &EntityOperationV1{
				TerraformActions:       []EntityOperationV1Config{},
				TerraformDataResources: []EntityOperationV1Config{},
				TerraformEphemeralResources: []EntityOperationV1Config{
					{
						Entity:         "Token",
						OperationTypes: []string{"open"},
						Order:          nil,
						Options:        nil,
					},
					{
						Entity:         "Secret",
						OperationTypes: []string{"open"},
						Order:          ptr(1),
						Options: &EntityOperationV1Options{
							Polling: &EntityOperationV1Polling{
								Name: "SecretPolling",
							},
						},
					},
				},
				TerraformManagedResources: []EntityOperationV1Config{},
			},
			wantErr: false,
		},
		{
			name: "mapping node - terraform-ephemeral-resource with null value",
			yaml: `
terraform-resource: User#create
terraform-ephemeral-resource: null`,
			expected: &EntityOperationV1{
				TerraformActions:            []EntityOperationV1Config{},
				TerraformDataResources:      []EntityOperationV1Config{},
				TerraformEphemeralResources: []EntityOperationV1Config{},
				TerraformManagedResources: []EntityOperationV1Config{
					{
						Entity:         "User",
						OperationTypes: []string{"create"},
						Order:          nil,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "scalar node - action invoke operation",
			yaml: `Task#invoke`,
			expected: &EntityOperationV1{
				TerraformActions: []EntityOperationV1Config{
					{
						Entity:         "Task",
						OperationTypes: []string{"invoke"},
						Order:          nil,
					},
				},
				TerraformDataResources:      []EntityOperationV1Config{},
				TerraformEphemeralResources: []EntityOperationV1Config{},
				TerraformManagedResources:   []EntityOperationV1Config{},
			},
			wantErr: false,
		},
		{
			name: "scalar node - action with order",
			yaml: `Task#invoke#1`,
			expected: &EntityOperationV1{
				TerraformActions: []EntityOperationV1Config{
					{
						Entity:         "Task",
						OperationTypes: []string{"invoke"},
						Order:          ptr(1),
					},
				},
				TerraformDataResources:      []EntityOperationV1Config{},
				TerraformEphemeralResources: []EntityOperationV1Config{},
				TerraformManagedResources:   []EntityOperationV1Config{},
			},
			wantErr: false,
		},
		{
			name: "sequence node - multiple actions",
			yaml: `
- Task#invoke
- Cleanup#invoke#1`,
			expected: &EntityOperationV1{
				TerraformActions: []EntityOperationV1Config{
					{
						Entity:         "Task",
						OperationTypes: []string{"invoke"},
						Order:          nil,
					},
					{
						Entity:         "Cleanup",
						OperationTypes: []string{"invoke"},
						Order:          ptr(1),
					},
				},
				TerraformDataResources:      []EntityOperationV1Config{},
				TerraformEphemeralResources: []EntityOperationV1Config{},
				TerraformManagedResources:   []EntityOperationV1Config{},
			},
			wantErr: false,
		},
		{
			name: "mapping node - terraform-action only",
			yaml: `
terraform-action: Task#invoke`,
			expected: &EntityOperationV1{
				TerraformActions: []EntityOperationV1Config{
					{
						Entity:         "Task",
						OperationTypes: []string{"invoke"},
						Order:          nil,
					},
				},
				TerraformDataResources:      []EntityOperationV1Config{},
				TerraformEphemeralResources: []EntityOperationV1Config{},
				TerraformManagedResources:   []EntityOperationV1Config{},
			},
			wantErr: false,
		},
		{
			name: "mapping node - terraform-action with sequence",
			yaml: `
terraform-action:
  - Task#invoke
  - Cleanup#invoke#2`,
			expected: &EntityOperationV1{
				TerraformActions: []EntityOperationV1Config{
					{
						Entity:         "Task",
						OperationTypes: []string{"invoke"},
						Order:          nil,
					},
					{
						Entity:         "Cleanup",
						OperationTypes: []string{"invoke"},
						Order:          ptr(2),
					},
				},
				TerraformDataResources:      []EntityOperationV1Config{},
				TerraformEphemeralResources: []EntityOperationV1Config{},
				TerraformManagedResources:   []EntityOperationV1Config{},
			},
			wantErr: false,
		},
		{
			name: "sequence node - action with mapping form and options",
			yaml: `
- entityOperation: Task#invoke
  options:
    polling:
      name: TaskPolling
      delaySeconds: 1
      intervalSeconds: 3
      limitCount: 10`,
			expected: &EntityOperationV1{
				TerraformActions: []EntityOperationV1Config{
					{
						Entity:         "Task",
						OperationTypes: []string{"invoke"},
						Order:          nil,
						Options: &EntityOperationV1Options{
							Polling: &EntityOperationV1Polling{
								Name:            "TaskPolling",
								DelaySeconds:    ptr(1),
								IntervalSeconds: ptr(3),
								LimitCount:      ptr(10),
							},
						},
					},
				},
				TerraformDataResources:      []EntityOperationV1Config{},
				TerraformEphemeralResources: []EntityOperationV1Config{},
				TerraformManagedResources:   []EntityOperationV1Config{},
			},
			wantErr: false,
		},
		{
			name: "mapping node - terraform-action with mapping form in sequence",
			yaml: `
terraform-action:
  - Task#invoke
  - entityOperation: Cleanup#invoke#1
    options:
      polling:
        name: CleanupPolling`,
			expected: &EntityOperationV1{
				TerraformActions: []EntityOperationV1Config{
					{
						Entity:         "Task",
						OperationTypes: []string{"invoke"},
						Order:          nil,
						Options:        nil,
					},
					{
						Entity:         "Cleanup",
						OperationTypes: []string{"invoke"},
						Order:          ptr(1),
						Options: &EntityOperationV1Options{
							Polling: &EntityOperationV1Polling{
								Name: "CleanupPolling",
							},
						},
					},
				},
				TerraformDataResources:      []EntityOperationV1Config{},
				TerraformEphemeralResources: []EntityOperationV1Config{},
				TerraformManagedResources:   []EntityOperationV1Config{},
			},
			wantErr: false,
		},
		{
			name: "mapping node - terraform-action with null value",
			yaml: `
terraform-resource: User#create
terraform-action: null`,
			expected: &EntityOperationV1{
				TerraformActions:            []EntityOperationV1Config{},
				TerraformDataResources:      []EntityOperationV1Config{},
				TerraformEphemeralResources: []EntityOperationV1Config{},
				TerraformManagedResources: []EntityOperationV1Config{
					{
						Entity:         "User",
						OperationTypes: []string{"create"},
						Order:          nil,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "scalar node - ephemeral resource close operation",
			yaml: `Token#close`,
			expected: &EntityOperationV1{
				TerraformActions:       []EntityOperationV1Config{},
				TerraformDataResources: []EntityOperationV1Config{},
				TerraformEphemeralResources: []EntityOperationV1Config{
					{
						Entity:         "Token",
						OperationTypes: []string{"close"},
						Order:          nil,
					},
				},
				TerraformManagedResources: []EntityOperationV1Config{},
			},
			wantErr: false,
		},
		{
			name: "scalar node - ephemeral resource close operation with order",
			yaml: `Token#close#1`,
			expected: &EntityOperationV1{
				TerraformActions:       []EntityOperationV1Config{},
				TerraformDataResources: []EntityOperationV1Config{},
				TerraformEphemeralResources: []EntityOperationV1Config{
					{
						Entity:         "Token",
						OperationTypes: []string{"close"},
						Order:          ptr(1),
					},
				},
				TerraformManagedResources: []EntityOperationV1Config{},
			},
			wantErr: false,
		},
		{
			name: "sequence node - ephemeral resource open and close operations",
			yaml: `
- Token#open
- Token#close`,
			expected: &EntityOperationV1{
				TerraformActions:       []EntityOperationV1Config{},
				TerraformDataResources: []EntityOperationV1Config{},
				TerraformEphemeralResources: []EntityOperationV1Config{
					{
						Entity:         "Token",
						OperationTypes: []string{"open"},
						Order:          nil,
					},
					{
						Entity:         "Token",
						OperationTypes: []string{"close"},
						Order:          nil,
					},
				},
				TerraformManagedResources: []EntityOperationV1Config{},
			},
			wantErr: false,
		},
		{
			name: "mapping node - terraform-ephemeral-resource close only",
			yaml: `
terraform-ephemeral-resource: Token#close`,
			expected: &EntityOperationV1{
				TerraformActions:       []EntityOperationV1Config{},
				TerraformDataResources: []EntityOperationV1Config{},
				TerraformEphemeralResources: []EntityOperationV1Config{
					{
						Entity:         "Token",
						OperationTypes: []string{"close"},
						Order:          nil,
					},
				},
				TerraformManagedResources: []EntityOperationV1Config{},
			},
			wantErr: false,
		},
		{
			name: "mapping node - terraform-ephemeral-resource with open and close in sequence",
			yaml: `
terraform-ephemeral-resource:
  - Token#open
  - Token#close#1`,
			expected: &EntityOperationV1{
				TerraformActions:       []EntityOperationV1Config{},
				TerraformDataResources: []EntityOperationV1Config{},
				TerraformEphemeralResources: []EntityOperationV1Config{
					{
						Entity:         "Token",
						OperationTypes: []string{"open"},
						Order:          nil,
					},
					{
						Entity:         "Token",
						OperationTypes: []string{"close"},
						Order:          ptr(1),
					},
				},
				TerraformManagedResources: []EntityOperationV1Config{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var node yaml.Node
			err := yaml.Unmarshal([]byte(tt.yaml), &node)
			require.NoError(t, err)

			// The root node is a document node, we need the content node
			require.Len(t, node.Content, 1, "Expected exactly one content node")
			contentNode := node.Content[0]

			e := &Extensions{}
			result, err := e.parseEntityOperationV1(contentNode)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
