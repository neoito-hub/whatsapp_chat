package general

import (
	"time"

	"github.com/lib/pq"
	"gopkg.in/go-playground/validator.v9"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	IDTOKEN             = "IDTOKEN"
	ACCESS              = "ACCESS"
	REFRESH             = "REFRESH"
	DEVICEACCESS        = "DEVICEACCESS"
	APPBACCESS          = "APPBACCESS"
	APPBREFRESH         = "APPBREFRESH"
	APPACCESS           = "APPACCESS"
	APPREFRESH          = "APPREFRESH"
	VERIFY_EMAIL_SECRET = "VERIFY_EMAIL_SECRET"

	APPBLOCKCLIENTID = "appblocks-9472"

	PROVIDER_SHIELD   = 1
	PROVIDER_GOOGLE   = 2
	PROVIDER_TWITTER  = 3
	PROVIDER_LINKEDIN = 4
)

var validate *validator.Validate

// type Envs struct {
// 	RedisHost     string
// 	RedisPort     string
// 	RedisPassword string
// 	RedisDB       string

// 	ClientSecretKey string

// 	GoogleClientId     string
// 	GoogleClientSecret string
// 	GoogleRedirectUrl  string

// 	LinkedInClientId     string
// 	LinkedInClientSecret string
// 	LinkedInRedirectUrl  string

// 	TwitterAPIKey    string
// 	TwitterAPISecret string
// 	TwitterCallback  string

// 	MailerEmail    string
// 	MailerPassword string
// 	MailerHost     string
// 	MailerPort     string

// 	IDTokenExpiry         string
// 	RefreshTokenExpiry    string
// 	AccessTokenExpiry     string
// 	AppAccessTokenExpiry  string
// 	AppRefreshTokenExpiry string

// 	ShieldPostgresHost     string
// 	ShieldPostgresUser     string
// 	ShieldPostgresPassword string
// 	ShieldPostgresName     string
// 	ShieldPostgresPort     string
// 	ShieldPostgresSslmode  string
// 	ShieldPostgresTimezone string
// }

// swagger:model Response
type ResponseTemplate struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Status  int         `json:"status"`
	Data    interface{} `json:"data"`
}

// swagger:parameters AppRegistrationRequest
//
//	in: body
type ApprRegEvent struct {
	AppId  string `json:"app_id" validate:"required"`
	UserID string `json:"user_id" validate:"required"`
}

type BlockApprRegEvent struct {
	AppId string `json:"app_id" validate:"required"`
}

type BlockApprRedirectUrl struct {
	AppId       string   `json:"app_id" validate:"required"`
	RedirectUrl []string `json:"redirect_url" validate:"required"`
}

// Model for Apps Table
type ShieldApp struct {
	//ID           uint
	AppId        string         `gorm:"primary_key; size:255"`
	ClientId     string         `gorm:"size:60; not null; unique"`
	ClientSecret string         `gorm:"size:255; not null"`
	UserId       string         `gorm:"size:255"`
	AppName      string         `gorm:"size:100; not null; unique"`
	AppSname     string         `gorm:"size:50"`
	Description  string         `gorm:"size:255"`
	LogoUrl      string         `gorm:"size:255"`
	AppUrl       string         `gorm:"size:255; not null"`
	RedirectUrl  pq.StringArray `gorm:"type:text[]; size:255; not null"`
	AppType      int            `gorm:"default:4"` // 1 - appblocks, 2 - internal app, 3 - client app, 4 - appblocks app
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Model for ShieldAppDomainMapping Table
type ShieldAppDomainMapping struct {
	//ID           uint
	ID         string `gorm:"primary_key; size:255"`
	OwnerAppID string
	Url        string

	OwnerApp ShieldApp `gorm:"foreignKey:OwnerAppID;References:AppId"`
}

// Model for Permissions Table
type Permission struct {
	PermissionId   string `gorm:"primary_key; size:255"`
	PermissionName string `gorm:"size:100; not null; unique"`
	Description    string `gorm:"size:255"`
	Category       string `gorm:"size:255"`
	Mandatory      bool   //to identify default mandatory permissions
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type GetScopes struct {
	PermissionId   string `json:"scope_id"`
	PermissionName string `json:"scope_name"`
	Editable       bool   `json:"editable"`
	Mandatory      bool   `json:"mandatory"`
	Selected       bool   `json:"selected"`
}

type CreateBlockAppScopes struct {
	AppId  string           `json:"app_id" validate:"required"`
	Scopes []BlockAppScopes `json:"scopes" validate:"required"`
}

type BlockAppScopes struct {
	PermissionId string `json:"scope_id" validate:"required"`
	Mandatory    bool   `json:"mandatory" validate:"required"`
}

// swagger:parameters PermissionIds
//
//	in: body
type PermissionIds struct {
	PermissionId string `json:"id"`
}

// Model for AppPermissions Table
type AppPermission struct {
	AppId        string `gorm:"primary_key; size:255; not null"`
	PermissionId string `gorm:"primary_key; size:255; not null"`
	Mandatory    bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// swagger:parameters AppPermissionsRequest
//
//	in: body
type AppPermissionEvent struct {
	AppId        string   `json:"appid" validate:"required"`
	PermissionId []string `json:"permissionid" validate:"required"`
	Mandatory    []bool   `json:"mandatory" validate:"required"`
}

// Model for AppUserPermissions Table
type AppUserPermission struct {
	UserId       string `gorm:"primary_key; size:255; not null"`
	AppId        string `gorm:"primary_key; size:255; not null"`
	PermissionId string `gorm:"primary_key; size:255; not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type AppUserPermissionEvent struct {
	AppId        string   `json:"appid" validate:"required"`
	PermissionId []string `json:"permissionid" validate:"required"`
}

type AppFromClientId struct {
	AppId       string
	ClientId    string
	RedirectUrl []string
	AppSname    string
	AppType     int
}

// swagger:parameters AppUrlParams
type AppUrlParams struct {
	// in: query
	// example: "appsname1-9651"
	ClientId string `json:"client_id"`
	// in: query
	// example: "code"
	ResponseType string `json:"response_type"`
	// in: query
	// example: "xyz"
	State string `json:"state"`
	// in: query
	// example: "http://localhost:8001"
	RedirectUri string `json:"redirect_uri"`
	GrantType   string
}

type AppPermissionsForAllow struct {
	PermissionId   string `json:"id"`
	PermissionName string `json:"name"`
	Description    string `json:"description"`
	Mandatory      bool   `json:"mandatory"`
}

type BlockAppCreds struct {
	ClientId     string    `json:"client_id"`
	ClientSecret string    `json:"client_secret"`
	CreatedAt    time.Time `json:"last_used"`
	UpdatedAt    time.Time `json:"created"`
}

type BlockAppClientIdResponse struct {
	ClientId  string    `json:"client_id"`
	CreatedAt time.Time `json:"last_used"`
	UpdatedAt time.Time `json:"created"`
}

type BlockAppClientSecretResponse struct {
	ClientSecret string    `json:"client_secret"`
	CreatedAt    time.Time `json:"last_used"`
	UpdatedAt    time.Time `json:"created"`
}

type TokenPairDetails struct {
	AccessToken  string
	RefreshToken string
	AccessUuid   string
	RefreshUuid  string
	AtExpires    int64
	RtExpires    int64
}

type TokenDetails struct {
	Token   string
	Uuid    string
	Expires int64
}

type UserAccessDetails struct {
	PairTokenUuid string
	UserId        string
	DeviceId      string
	Refreshed     bool
}

type AppAccessDetails struct {
	PairTokenUuid     string
	UserId            string
	ClientId          string
	AppSname          string
	AppUserPermission []string
	DeviceId          string
	Refreshed         bool
}

type AccessDetails struct {
	AccessUuid  string
	RefreshUuid string
	UserId      string
	DeviceId    string
}

type TokenPairs struct {
	AccessToken  string
	RefreshToken string
}

type AuthCodeValues struct {
	UserId            string
	ClientId          string
	AppSname          string
	AppUserPermission []string
	CodeName          string
	DeviceId          string
}

// swagger:parameters SignupRequest
type SignupRequestEvent struct {
	// in: body
	// example: "email1@email.com"
	Email string `json:"email" validate:"required,email"`
	// in: body
	// example: "@ppBlock1"
	Password string `json:"password" validate:"required"`
	// in: body
	// example: "username1"
	UserName string `json:"username" validate:"required"`
	// in: body
	// example: "Full Name"
	FullName string `json:"fullname"`
	// in: body
	// example: "address line 1"
	Address1 string `json:"address1" validate:"required_with=Address2"`
	// in: body
	// example: "address line 2"
	Address2 string `json:"address2"`
	// in: body
	// example: "9876543210"
	Phone string `json:"phone"`
}

// Model for User Table
type User struct {
	UserID                  string `gorm:"primaryKey; size:255"` // FK to MEMBER table (type U).
	UserName                string `gorm:"size:255; not null; unique"`
	FullName                string `gorm:"size:255"`
	Email                   string `gorm:"size:255; not null; unique"`
	Password                string `gorm:"size:255; not null"`
	Address1                string `gorm:"size:150"`
	Address2                string `gorm:"size:150"`
	Phone                   string `gorm:"size:20"`
	EmailVerificationCode   string `gorm:"size:6"`
	EmailVerified           bool   `gorm:"default:false"`
	EmailVerificationExpiry time.Time
	CreatedAt               time.Time
	UpdatedAt               time.Time
	OptCounter              int `gorm:"size:8"`

	UserMember Member `gorm:"foreignKey:UserID;References:ID"`
}

type Member struct {
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
	ID         string         `gorm:"primaryKey;not null"`
	Type       string         `gorm:"size:3"` //  (S-space, U-user, T-teams)
	OptCounter int            `gorm:"size:8"`
}

type Space struct {
	CreatedAt             time.Time
	UpdatedAt             time.Time
	DeletedAt             gorm.DeletedAt `gorm:"index"`
	SpaceID               string         `gorm:"primaryKey;not null"` // NOT NULL	FK to MEMBER table for this space or space unit (type S).
	LegalID               string         // The registered space identifier, given to the space (such as assigned by the government). This may be null for an space unit. This is not the name of the space, which should be stored in the ORGENTITYNAME table.
	Type                  string         `gorm:"not null"`             // NOT NULL P -> personal, B -> business or institution
	Name                  string         `gorm:"unique; not null"`     // Name of the space
	BusinessName          string         `gorm:"unique; default:null"` // The business name of the space. (Unique if value exist)
	Address               string         // Address of the business or institution
	Email                 string         // Email of the space (Unique if value exist)
	Country               string         // Country of the space (Not null for business or institution)
	BusinessCategory      string         // The business category, which describes the kind of business performed by an Space.
	Description           string         // A description of the Space.
	MetaData              datatypes.JSON // Additional metadata about the Space.
	TaxPayerID            string         // A string used to identify the Space for taxation purpose. Addition of this column triggered by Taxware integration, but presumably this column is useful even outside of Taxware.
	DistinguishedName     string         // Distinguished name (DN) of the Space. If LDAP is used, contains the DN of the Space in the LDAP server. If database is used as member repository, contains a unique name as defined by the membership hierarchy. DNs for all OrgEntities are logically unique, however due to the large field size, this constraint is not enforced at the physical level. The DN should not contain any spaces before or after the comma (,) or equals sign (=) and must be entered in lowercase.
	Status                int            // NOT NULL DEFAULT 0	The STATUS column in Space table indicates whether or not the space is locked. Valid values are as follows: 0 = not locked -1 = locked
	OptCount              int            //	The optimistic concurrency control counter for the table. Every time there is an update to the table, the counter is incremented.
	MarketPlaceID         string         // The ID of the market place where the space is registered.
	DeveloperPortalAccess bool           `gorm:"not null;default:false"` // Indicates whether the space has access to the developer portal.
	OptCounter            int            `gorm:"size:8"`

	Member Member `gorm:"foreignKey:SpaceID;References:ID;unique;not null"`
	// TODO add reference to marketplace

}

type Role struct {
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
	ID           string         `gorm:"primaryKey;not null"`
	Name         string
	Description  string
	OwnerSpaceID string
	IsOwner      bool
	CreatedBy    string
	UpdatedBy    string
	OptCounter   int `gorm:"size:8"`

	OwnerSpace Space `gorm:"foreignKey:OwnerSpaceID;References:SpaceID"`
}

type MemberRole struct {
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
	ID           string         `gorm:"primaryKey;not null"`
	OwnerUserID  string         // PK and FK to MEMBER table (type U).
	RoleID       string         `gorm:"not null"` // NOT NULL	FK to Role.
	OwnerSpaceID string         // PK and FK to MEMBER table (type U).
	OptCounter   int            `gorm:"size:8"`

	MemberUser  User  `gorm:"foreignKey:OwnerUserID;references:UserID"`
	Role        Role  `gorm:"foreignKey:RoleID;references:ID"`
	SpaceMember Space `gorm:"foreignKey:OwnerSpaceID;references:SpaceID"`
}

type SpaceMember struct {
	gorm.Model

	ID           string `gorm:"primaryKey;not null"`
	OwnerUserID  string `gorm:"index:member_role_unique_index,unique"` // PK and FK to MEMBER table (type U).
	OwnerSpaceID string `gorm:"index:member_role_unique_index,unique"` // PK and FK to MEMBER table (type U).
	OptCounter   int    `gorm:"size:8"`

	MemberUser  User  `gorm:"foreignKey:OwnerUserID;references:UserID"`
	SpaceMember Space `gorm:"foreignKey:OwnerSpaceID;references:SpaceID"`
}

type DefaultUserSpace struct {
	gorm.Model

	ID           string `gorm:"primaryKey"`
	OwnerUserID  string `gorm:"unique"` // FK to User table
	OwnerSpaceID string // FK to Spaces table
	OptCounter   int    `gorm:"size:8"`
	OwnerUser    User   `gorm:"foreignKey:OwnerUserID;References:UserID"`
	OwnerSpace   Space  `gorm:"foreignKey:OwnerSpaceID;References:SpaceID"`
}

type Team struct {
	TeamID      string `gorm:"primaryKey;not null"` // PK and FK to MEMBER table (type T).
	OwnerID     string `gorm:"not null"`            // PK and FK to MEMBER table  (type S).
	Name        string
	Description string
	Update      string
	UpdatedBy   string
	OptCounter  int `gorm:"size:8"`

	Member      Member `gorm:"foreignKey:TeamID;References:ID"`
	OwnerMember Member `gorm:"foreignKey:OwnerID;References:ID"`
}

type TeamMember struct {
	ID          string `gorm:"primaryKey;not null"`
	OwnerTeamID string // PK and FK to Team.
	MemberID    string // PK and FK to MEMBER table (type U).
	OptCounter  int    `gorm:"size:8"`

	Team   Team   `gorm:"foreignKey:OwnerTeamID;References:TeamID"`
	Member Member `gorm:"foreignKey:MemberID;References:ID"`
}

type AcResource struct {
	ID          string `gorm:"primaryKey;not null"`
	Name        string
	Description string
	Path        string
	OptCounter  int `gorm:"size:8"`
}

type AcResGrp struct {
	ID           string `gorm:"primaryKey;not null"`
	MemberID     string // FK to MEMBER table
	Description  string
	IsPredefined bool
	OptCounter   int `gorm:"size:8"`

	OwnerMember Member `gorm:"foreignKey:MemberID;References:ID"`
}

type AcResGpRes struct {
	ID           string `gorm:"primaryKey;not null"`
	AcResGrpID   string
	AcResourceID string // FK to MEMBER table
	OptCounter   int    `gorm:"size:8"`

	AcResource AcResource `gorm:"foreignKey:AcResourceID;references:ID"`
	AcResGrp   AcResGrp   `gorm:"foreignKey:AcResGrpID;references:ID"`
}

type AcResAction struct {
	ID           string `gorm:"primaryKey;not null"`
	AcActionID   string
	AcResourceID string // FK to MEMBER table
	OptCounter   int    `gorm:"size:8"`

	AcResource AcResource `gorm:"foreignKey:AcResourceID;References:ID"`
	AcAction   AcAction   `gorm:"foreignKey:AcActionID;References:ID"`
}

type AcAction struct {
	ID          string `gorm:"primaryKey;not null"`
	Name        string
	Description string
	OptCounter  int `gorm:"size:8"`
}

type AcActGrp struct {
	ID           string `gorm:"primaryKey;not null"`
	MemberID     string // FK to MEMBER table
	Description  string
	IsPredefined bool
	OptCounter   int `gorm:"size:8"`

	OwnerMember Member `gorm:"foreignKey:MemberID;References:ID"`
}

type ActGpAction struct {
	ID         string `gorm:"primaryKey"`
	AcActGrpID string
	AcActionID string
	OptCounter int `gorm:"size:8"`

	AcAction AcAction `gorm:"foreignKey:AcActionID;References:ID"`
	AcActGrp AcActGrp `gorm:"foreignKey:AcActGrpID;References:ID"`
}

type AcPolicy struct {
	ID           string `gorm:"primaryKey;not null"`
	AcActGrpID   string
	AcResGrpID   string
	MemberID     string // FK to MEMBER table
	CreatedBy    string // FK to USERS TABLE
	UpdatedBy    string // FK TO USERS TABLE
	Name         string
	Description  string
	Path         string
	OptCounter   int `gorm:"size:8"`
	IsPredefined bool

	AcActionGroup   AcActGrp `gorm:"foreignKey:AcActGrpID;References:ID"`
	AcResourceGroup AcResGrp `gorm:"foreignKey:AcResGrpID;References:ID"`
	OwnerMember     Member   `gorm:"foreignKey:MemberID;References:ID"`
	CreatedUser     User     `gorm:"foreignKey:CreatedBy;References:UserID"`
	UpdatedUser     User     `gorm:"foreignKey:UpdatedBy;References:UserID"`
}

type AcPolGrp struct {
	ID          string `gorm:"primaryKey;not null"`
	MemberID    string `gorm:"not null"` // FK to MEMBER table
	Description string `gorm:"size:255; not null"`
	OptCounter  int    `gorm:"size:8"`

	OwnerMember Member `gorm:"foreignKey:MemberID;References:ID"`
}

type PolGpPolicy struct {
	ID         string `gorm:"primaryKey;not null"`
	AcPolicyID string
	AcPolGrpID string // FK to MEMBER table
	OptCounter int    `gorm:"size:8"`

	AcPolGrp AcPolGrp `gorm:"foreignKey:AcPolGrpID;References:ID"`
	AcPolicy AcPolicy `gorm:"foreignKey:AcPolicyID;References:ID"`
}

type AcPolGrpSub struct {
	ID           string `gorm:"primaryKey;not null"`
	OwnerSpaceID string // FK to Spaces table
	RoleID       string // FK to Role table
	OwnerTeamID  string // FK to Team table
	AcPolGrpID   string // FK to AcPolGrp table
	OptCounter   int    `gorm:"size:8"`

	AcPolGrp   AcPolGrp `gorm:"foreignKey:AcPolGrpID;References:ID"`
	OwnerSpace Space    `gorm:"foreignKey:OwnerSpaceID;References:SpaceID"`
	OwnerRole  Role     `gorm:"foreignKey:RoleID;References:ID"`
	OwnerTeam  Team     `gorm:"foreignKey:OwnerTeamID;References:TeamID"`
}
type RolePayload struct {
	SpaceID string `json:"space_id"`
	RoleID  string `json:"role_id"`
}

type AcPolGrpData struct {
	ID string
}

type PolicyData struct {
	ID string
}

type PolGpPolicyData struct {
	ID string
}

type AcPolGrpSubData struct {
	ID     string
	RoleID string
}

type UserProvider struct {
	UserId    string `gorm:"primary_key; size:255; not null"`
	Provider  int    `gorm:"primary_key; not null; default:1"` // 1 - shield, 2 - google
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Model for User Details
type UserDetails struct {
	UserId   string         `json:"user_id"`
	Provider pq.StringArray `json:"provider" gorm:"type:text[]"`
	UserName string         `json:"user_name"`
	FullName string         `json:"full_name"`
	Email    string         `json:"email"`
	Address1 string         `json:"address1"`
	Address2 string         `json:"address2"`
	Phone    string         `json:"phone"`
}

// Model for User Details
type UserDetailsWithProvider struct {
	UserId   string `json:"user_id"`
	Provider string `json:"provider"`
	UserName string `json:"user_name"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Address1 string `json:"address1"`
	Address2 string `json:"address2"`
	Phone    string `json:"phone"`
}

// swagger:parameters LoginRequest
//
//	in: body
type LoginRequestEvent struct {
	EmailOrUserName string `json:"email" validate:"required"`
	Password        string `json:"password" validate:"required"`
}

type VerfiyEmailRequestEvent struct {
	UserId           string `json:"user_id" validate:"required,user_id"`
	VerificationCode string `json:"verification_code" validate:"required,verification_code"`
}

type ResendOTPRequestEvent struct {
	UserId string `json:"user_id" validate:"required"`
}

// GoogleUser is to handle the response body of google user info request.
type GoogleUser struct {
	Sub           string `json:"sub"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Profile       string `json:"profile"`
	Picture       string `json:"picture"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Gender        string `json:"gender"`
}

type SocialLoginDetails struct {
	ClientId     string `json:"client_id"`
	ResponseType string `json:"response_type"`
	State        string `json:"state"`
	RedirectUri  string `json:"redirect_uri"`
	GrantType    string `json:"grant_type"`
	IsLogin      bool   `json:"is_login"`
}

type ResetPasswordRequestEvent struct {
	Email string `json:"email" validate:"required"`
}

type ResetPasswordDetails struct {
	ClientId     string `json:"client_id"`
	ResponseType string `json:"response_type"`
	State        string `json:"state"`
	RedirectUri  string `json:"redirect_uri"`
	GrantType    string `json:"grant_type"`
	UserId       string `json:"user_id"`
}

type ResetPasswordUser struct {
	UserID   string
	Email    string
	UserName string
}

type ResetPasswordMail struct {
	Email     string
	UserName  string
	ResetLink string
}

type UserWithProvider struct {
	UserId        string
	UserName      string
	Email         string
	Password      string
	EmailVerified bool
	Provider      int // 1 - shield, 2 - google
}

type TwitterCreds struct {
	Key      string
	Sec      string
	Callback string
}

type LinkedinATResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type ChangePasswordRequestEvent struct {
	Password   string `json:"password" validate:"required"`
	PasswordRe string `json:"passwordRe" validate:"required"`
}

type ChangeUserPasswordRequestEvent struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required"`
}

type UpdateUserProfileRequestEvent struct {
	UserName string `json:"username"`
	FullName string `json:"fullname"`
	Address1 string `json:"address1"`
	Address2 string `json:"address2"`
	Phone    string `json:"phone"`
}

type ResetUserPasswordDetails struct {
	UserID   string
	Email    string
	UserName string
	Password string
}

type ActiveSession struct {
	IsActiveu_sid bool
	Sessionstring string
}
