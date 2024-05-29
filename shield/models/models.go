// UNION OF BLOCK REGISTRY (feat/block-discovery) model WITH PAYMENTS (develop) model (Jan 17 2023)

package models

import (
	"time"

	"github.com/lib/pq"

	// "github.com/shopspring/decimal"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// type Block struct {
// 	gorm.Model
// 	ID                 string `gorm:"primaryKey; not null"`
// 	BlockType          int    // app - 1 , ui-container - 2,ui-elements 3  fn - 4, data - 5,function shared block -6
// 	BlockName          string `gorm:"not null"`
// 	BlockShortName     string
// 	BlockDesc          string
// 	IsPublic           bool
// 	GitUrl             string
// 	Lang               int   // 0 - js, 1 - go
// 	Status             int   `gorm:"not null;default:1"`     // 1 - active, 2 - disabled, 3 - archived, 4 -  deleted.
// 	Verified           bool  `gorm:"not null;default:false"` // true - verified, false - non-verified
// 	BlockVisibility    int64 `gorm:"not null;default:3"`     // 1 public 2 private 3 paid 4 free 5 Temporary block
// 	RootPackageBlockID string
// 	IsSearchable       bool     `gorm:"default:false"`
// 	RootEntity         Entities `gorm:"foreignKey:RootPackageBlockID;References:EntityID"`
// 	Entity             Entities `gorm:"foreignKey:ID;References:EntityID;unique;not null"`
// }

// type CronLogs struct {
// 	gorm.Model
// 	ID   string       `gorm:"primaryKey; not null"`
// 	Data pgtype.JSONB `gorm:"type:jsonb"`
// 	Name string
// }

// type BlockVariantHierarchy struct {
// 	gorm.Model
// 	ID                string `gorm:"primaryKey; not null"`
// 	Ancestor          string // block id of ancestor
// 	Descendant        string // block id of descendant
// 	VersionId         *string
// 	NumLevel          int
// 	AncestorVariant   Block         `gorm:"ForeignKey:Ancestor;References:ID"`
// 	DescendantVariant Block         `gorm:"ForeignKey:Descendant;References:ID"`
// 	AncestorVersion   BlockVersions `gorm:"ForeignKey:VersionId;References:ID"`
// }

// type BlockMetaData struct {
// 	CreatedAt      time.Time
// 	UpdatedAt      time.Time
// 	DeletedAt      gorm.DeletedAt `gorm:"index"`
// 	BlockId        string
// 	BlockVersionID string
// 	Schema         datatypes.JSON // only for fns
// 	SampleEnv      datatypes.JSON // only for fns
// 	AppConfig      datatypes.JSON // only for app
// 	ReadMe         string         // avail for all blocks
// 	ParentBlockIDs pq.StringArray `gorm:"type:text[]"`

// 	Block        Block         `gorm:"ForeignKey:BlockId;References:ID"`
// 	BlockVersion BlockVersions `gorm:"ForeignKey:BlockVersionID;References:ID"`
// }

// public private - 0
// public - 1
// private - 2

// type BlockMapping struct {
// 	gorm.Model
// 	ID             string `gorm:"primaryKey; not null"`
// 	AppBlockId     string
// 	RelationType   int  // dependency - 1 or composition - 2
// 	IsAPI          bool // ui - fn - true
// 	BlockId        string
// 	RelatedBlockId string
// 	Status         int   `gorm:"not null;default:1"` // 1 - active, 2 - disabled, 3 - archived, 4 -  deleted.
// 	AppBlock       Block `gorm:"ForeignKey:AppBlockId;References:ID"`
// 	Block          Block `gorm:"ForeignKey:BlockId;References:ID"`
// 	RelatedBlock   Block `gorm:"ForeignKey:RelatedBlockId;References:ID"`
// }

// type Tags struct {
// 	gorm.Model
// 	ID      string `gorm:"primaryKey; not null"`
// 	TagName string
// }

// type Category struct {
// 	gorm.Model
// 	ID           string `gorm:"primaryKey; not null"`
// 	CategoryName string
// 	Description  string
// }

// type BlockCategoryMapping struct {
// 	gorm.Model
// 	ID            string `gorm:"primaryKey; not null"`
// 	CategoryID    string
// 	BlockID       string
// 	Block         Block    `gorm:"ForeignKey:BlockID;References:ID"`
// 	BlockCategory Category `gorm:"ForeignKey:CategoryID;References:ID"`
// }

// type SDMModel struct {
// 	gorm.Model
// 	ID         string         `gorm:"primaryKey; not null"`
// 	Name       string         `gorm:"not null; UniqueIndex:sdm_name_version"`
// 	Version    string         `gorm:"not null; UniqueIndex:sdm_name_version"`
// 	SchemaJSON datatypes.JSON `gorm:"not null"`
// 	//
// }

// type SDMModelSDMVersionMapping struct {
// 	gorm.Model
// 	ID           string `gorm:"primaryKey; not null"`
// 	SDMModelID   string `gorm:"not null"`
// 	SDMVersionID string `gorm:"not null"`

// 	SDMVersion SDMVersion `gorm:"ForeignKey:SDMVersionID;References:ID"`
// 	SDMModel   SDMModel   `gorm:"ForeignKey:SDMModelID;References:ID"`
// }

// type SDMModelBusinessDomainMapping struct {
// 	gorm.Model
// 	ID               string `gorm:"primaryKey; not null"`
// 	SDMModelID       string `gorm:"not null"`
// 	BusinessDomainID string `gorm:"not null"`

// 	SDMModel       SDMModel       `gorm:"ForeignKey:SDMModelID;References:ID"`
// 	BusinessDomain BusinessDomain `gorm:"ForeignKey:BusinessDomainID;References:ID"`
// }
// type BusinessDomain struct {
// 	gorm.Model
// 	ID          string `gorm:"primaryKey; not null"`
// 	Name        string `gorm:"not null; unique"`
// 	DisplayName string `gorm:"not null;"`
// 	Description string
// }

// type SDMModelSubmit struct {
// 	gorm.Model
// 	ID                string         `gorm:"primaryKey; not null"`
// 	Name              string         `gorm:"not null"`
// 	SDMVersionIDs     pq.StringArray `gorm:"type:text[];not null"`
// 	SchemaJSON        datatypes.JSON `gorm:"not null"`
// 	BusinessDomainIDs pq.StringArray `gorm:"type:text[];not null"`
// 	SubmittedBy       string
// 	SDMModelID        string    `gorm:"default:null"`
// 	ReviewedBy        string    `gorm:"default:null"`
// 	ReviewedOn        time.Time `gorm:"type:timestamp"`
// 	Status            int       `gorm:"default:0"` // 0 pending, 1 approved, 2 rejected
// 	//

// 	SDMModel      SDMModel `gorm:"ForeignKey:SDMModelID;References:ID"`
// 	SubmittedUser User     `gorm:"ForeignKey:SubmittedBy;References:UserID"`
// 	ReviewedUser  User     `gorm:"ForeignKey:ReviewedBy;References:UserID"`
// }

// type CategoryHierarchy struct {
// 	gorm.Model
// 	ID                 string `gorm:"primaryKey; not null"`
// 	Ancestor           string
// 	Descendant         string
// 	NumLevel           int
// 	AncestorCategory   Category `gorm:"ForeignKey:Ancestor;References:ID"`
// 	DescendantCategory Category `gorm:"ForeignKey:Descendant;References:ID"`
// }

// type BlockTagsMapping struct {
// 	gorm.Model
// 	ID       string `gorm:"primaryKey; not null"`
// 	TagId    string
// 	BlockId  string
// 	Block    Block `gorm:"ForeignKey:BlockId;References:ID"`
// 	BlockTag Tags  `gorm:"ForeignKey:TagId;References:ID"`
// }

// type BlockVersions struct {
// 	gorm.Model
// 	ID            string `gorm:"primaryKey"`
// 	VersionNumber string
// 	BlockId       string
// 	// IsRelease     bool
// 	Status        int // 1 - in_dev, 2 - release_ready, 3 - reject, 4 - approved, 5 - archive, 6 - pending approval, 7 - un-discoverable
// 	ReleaseNotes  string
// 	SourceCodeKey string
// 	IsPreviewable bool
// 	BlockTypesMap pgtype.JSONB `gorm:"type:jsonb"`
// 	Block         Block        `gorm:"ForeignKey:BlockId;References:ID"`
// }

// type BlockAppAssign struct {
// 	gorm.Model
// 	ID      string `gorm:"primaryKey"`
// 	AppID   string `gorm:"not null; uniqueIndex:baa_app_id_block_id"`
// 	BlockID string `gorm:"not null; uniqueIndex:baa_app_id_block_id"`
// 	SpaceID string `gorm:"not null;"`
// 	Status  int    // 1 - assigned, 2 - in use

// 	Block Block `gorm:"ForeignKey:BlockID;References:ID"`
// 	App   App   `gorm:"ForeignKey:AppID;References:ID"`
// 	Space Space `gorm:"ForeignKey:SpaceID;References:ID"`
// }
// type BlockAuthors struct {
// 	gorm.Model
// 	UserId     string
// 	BlockId    string
// 	AuthorType int   // 1 - Owner, 2 - Admin, 3 - Collabrator
// 	Block      Block `gorm:"ForeignKey:BlockId;References:ID"`
// }

// owner can invite anyone
// owner can  change ownership to anyone else
// Admin can invite other admin?
// admin can invite collabrator

// type BlockActivity struct {
// 	gorm.Model
// 	ID           string `gorm:"primaryKey"`
// 	BlockId      string
// 	UserId       string
// 	ActivityType string
// 	Block        Block `gorm:"ForeignKey:BlockId;References:ID"`
// }

// type Job struct {
// 	gorm.Model
// 	ID       string `gorm:"primaryKey"`
// 	BlockID  string `gorm:"not null"`
// 	Schedule string // corn job schedule time
// 	TimeZone string

// 	Block Block `gorm:"ForeignKey:BlockID;References:ID"`
// }

// type JobLogs struct {
// 	gorm.Model
// 	ID    string         `gorm:"primaryKey"`
// 	JobID string         `gorm:"not null"`
// 	log   datatypes.JSON // log data

// 	Job Job `gorm:"ForeignKey:JobID;References:ID"`
// }

// block space mapping table
// type BlockSpaceMapping struct {
// 	gorm.Model
// 	ID              string `gorm:"primaryKey; not null"`
// 	SpaceID         string `gorm:"not null"`
// 	SpaceName       string `gorm:"not null; UniqueIndex:bsm_space_name_block_name_root_package_name"`
// 	BlockID         string `gorm:"not null"`
// 	BlockName       string `gorm:"not null; UniqueIndex:bsm_space_name_block_name_root_package_name"`
// 	RootPackageName string `gorm:"; UniqueIndex:bsm_space_name_block_name_root_package_name"`

// 	Block Block `gorm:"ForeignKey:BlockID;References:ID"`
// 	Space Space `gorm:"ForeignKey:SpaceID;References:SpaceID"`
// }

// type AppblockVersion struct {
// 	gorm.Model
// 	ID          string `gorm:"primaryKey; not null"`
// 	Name        string
// 	Version     string `gorm:"unique; not null"`
// 	Description string
// }

// type Language struct {
// 	gorm.Model
// 	ID          string `gorm:"primaryKey; not null"`
// 	Name        string `gorm:"unique; not null"`
// 	Description string
// 	S3Url       string
// }

// TO BE REMOVED
type Runtime struct {
	gorm.Model
	ID      string `gorm:"primaryKey; not null"`
	Name    string `gorm:"not null; UniqueIndex:idx_name_version"`
	Version string `gorm:"not null; UniqueIndex:idx_name_version"`
}

// TO BE REMOVED
// type BlockRuntimeMapping struct {
// 	gorm.Model
// 	ID             string `gorm:"primaryKey; not null"`
// 	BlockID        string `gorm:"not null; UniqueIndex:idx_block_id_block_version_id_runtime_id"`
// 	BlockVersionID string `gorm:"not null; UniqueIndex:idx_block_id_block_version_id_runtime_id"`
// 	RuntimeID      string `gorm:"not null; UniqueIndex:idx_block_id_block_version_id_runtime_id"`

// 	Block   Block         `gorm:"ForeignKey:BlockID;References:ID"`
// 	Version BlockVersions `gorm:"ForeignKey:BlockVersionID;References:ID"`
// 	Runtime Runtime       `gorm:"ForeignKey:RuntimeID;References:ID"`
// }

// type LanguageVersion struct {
// 	gorm.Model
// 	ID         string `gorm:"primaryKey; not null"`
// 	Name       string `gorm:"not null; UniqueIndex:idx_name_version_type"`
// 	Version    string `gorm:"not null; UniqueIndex:idx_name_version_type"`
// 	Type       int    `gorm:"not null; uniqueIndex:idx_name_version_type; default:0"` // 0 - version, 1 - runtime_version
// 	LanguageID string `gorm:"not null"`

// 	Language Language `gorm:"ForeignKey:LanguageID;References:ID"`
// }

// type SDMVersion struct {
// 	gorm.Model
// 	ID      string `gorm:"primaryKey; not null"`
// 	Name    string `gorm:"not null;"`
// 	Version string `gorm:"not null; unique"`
// }

// type AppblockVersionSDMVersionMapping struct {
// 	gorm.Model
// 	ID                string `gorm:"primaryKey; not null"`
// 	AppblockVersionID string `gorm:"not null; UniqueIndex:avrm_appblock_version_id_sdm_version_id"`
// 	SDMVersionID      string `gorm:"not null; UniqueIndex:avrm_appblock_version_id_sdm_version_id"`

// 	SDMVersion      SDMVersion      `gorm:"ForeignKey:SDMVersionID;References:ID"`
// 	AppblockVersion AppblockVersion `gorm:"ForeignKey:AppblockVersionID;References:ID"`
// }

// type Dependency struct {
// 	gorm.Model
// 	ID          string `gorm:"primaryKey; not null"`
// 	Name        string `gorm:"not null;  UniqueIndex:dependencies_name_version_type"`
// 	Version     string `gorm:"not null;  UniqueIndex:dependencies_name_version_type"`
// 	Type        int    `gorm:"not null;  UniqueIndex:dependencies_name_version_type"` // 0 dep , 1 - devDep
// 	Description string
// 	URL         string
// }

// type DependencySubmit struct {
// 	gorm.Model
// 	ID                 string `gorm:"primaryKey; not null"`
// 	Name               string `gorm:"not null"`
// 	Version            string `gorm:"not null"`
// 	Type               int    `gorm:"not null"` // 0 dep , 1 - devDep
// 	URL                string
// 	Description        string
// 	Status             int            // 0 pending, 1 approved, 2 rejected
// 	LanguageVersionIDs pq.StringArray `gorm:"type:text[]; not null"`
// 	SubmittedBy        string         `gorm:"default:null"`
// 	ReviewedBy         string         `gorm:"default:null"`
// 	ReviewedOn         time.Time      `gorm:"type:timestamp"`

// 	SubmittedUser User `gorm:"ForeignKey:SubmittedBy;References:UserID"`
// 	ReviewedUser  User `gorm:"ForeignKey:ReviewedBy;References:UserID"`
// }

// type BlockDependencyMapping struct {
// 	gorm.Model
// 	ID             string `gorm:"primaryKey; not null"`
// 	BlockID        string `gorm:"not null; UniqueIndex:idx_block_id_block_version_id_dependency_id"`
// 	BlockVersionID string `gorm:"not null; UniqueIndex:idx_block_id_block_version_id_dependency_id"`
// 	DependencyID   string `gorm:"not null; UniqueIndex:idx_block_id_block_version_id_dependency_id"`

// 	Block      Block         `gorm:"ForeignKey:BlockID;References:ID"`
// 	Version    BlockVersions `gorm:"ForeignKey:BlockVersionID;References:ID"`
// 	Dependency Dependency    `gorm:"ForeignKey:DependencyID;References:ID"`
// }

// type BlockLanguageVersionMapping struct {
// 	gorm.Model
// 	ID                string `gorm:"primaryKey; not null"`
// 	BlockID           string `gorm:"not null; UniqueIndex:lvbm_block_id_block_version_id_language_version_id"`
// 	BlockVersionID    string `gorm:"not null; UniqueIndex:lvbm_block_id_block_version_id_language_version_id"`
// 	LanguageVersionID string `gorm:"not null; UniqueIndex:lvbm_block_id_block_version_id_language_version_id"`

// 	Block           Block           `gorm:"ForeignKey:BlockID;References:ID"`
// 	Version         BlockVersions   `gorm:"ForeignKey:BlockVersionID;References:ID"`
// 	LanguageVersion LanguageVersion `gorm:"ForeignKey:LanguageVersionID;References:ID"`
// }

// type BlockAppblockVersionMapping struct {
// 	gorm.Model
// 	ID                string `gorm:"primaryKey; not null"`
// 	BlockID           string `gorm:"not null; UniqueIndex:avbm_block_id_block_version_id_appblock_version_id"`
// 	BlockVersionID    string `gorm:"not null; UniqueIndex:avbm_block_id_block_version_id_appblock_version_id"`
// 	AppblockVersionID string `gorm:"not null; UniqueIndex:avbm_block_id_block_version_id_appblock_version_id"`

// 	Block           Block           `gorm:"ForeignKey:BlockID;References:ID"`
// 	Version         BlockVersions   `gorm:"ForeignKey:BlockVersionID;References:ID"`
// 	AppblockVersion AppblockVersion `gorm:"ForeignKey:AppblockVersionID;References:ID"`
// }

// type AppblockVersionLanguageVersionMapping struct {
// 	gorm.Model
// 	ID                string `gorm:"primaryKey; not null"`
// 	AppblockVersionID string `gorm:"not null; UniqueIndex:avrm_appblock_version_id_language_version_id"`
// 	LanguageVersionID string `gorm:"not null; UniqueIndex:avrm_appblock_version_id_language_version_id"`

// 	LanguageVersion LanguageVersion `gorm:"ForeignKey:LanguageVersionID;References:ID"`
// 	AppblockVersion AppblockVersion `gorm:"ForeignKey:AppblockVersionID;References:ID"`
// }

// type AppblockVersionLanguageMapping struct {
// 	gorm.Model
// 	ID                string `gorm:"primaryKey; not null"`
// 	AppblockVersionID string `gorm:"not null; UniqueIndex:avrm_appblock_version_id_language_id"`
// 	LanguageID        string `gorm:"not null; UniqueIndex:avrm_appblock_version_id_language_id"`

// 	Language        Language        `gorm:"ForeignKey:LanguageID;References:ID"`
// 	AppblockVersion AppblockVersion `gorm:"ForeignKey:AppblockVersionID;References:ID"`
// }

// type DependencyLanguageVersionMapping struct {
// 	gorm.Model
// 	ID                string `gorm:"primaryKey; not null"`
// 	DependencyID      string `gorm:"not null; UniqueIndex:dprm_dependency_id_language_version_id"`
// 	LanguageVersionID string `gorm:"not null; UniqueIndex:dprm_dependency_id_language_version_id"`

// 	LanguageVersion LanguageVersion `gorm:"ForeignKey:LanguageVersionID;References:ID"`
// 	Dependency      Dependency      `gorm:"ForeignKey:DependencyID;References:ID"`
// }

type PolGrpSubsEntityMapping struct {
	gorm.Model

	ID            string `gorm:"primaryKey;not null"`
	OwnerEntityID string
	PolGrpSubsID  string

	PolicyGroupSubs AcPolGrpSub `gorm:"foreignKey:PolGrpSubsID;References:ID"`
	Entities        Entities    `gorm:"foreignKey:OwnerEntityID;References:EntityID"`
}

type AcPolGrpSub struct {
	gorm.Model
	ID           string `gorm:"primaryKey;not null"`
	OwnerSpaceID string // FK to Spaces table
	RoleID       string `gorm:"default:null"` // FK to Role table
	OwnerTeamID  string `gorm:"default:null"` // FK to Team table
	OwnerUserID  string `gorm:"default:null"` // FK to User table
	AcPolGrpID   string // FK to AcPolGrp table
	OptCounter   int    `gorm:"size:8"`
	PermissionID string `gorm:"default:null"` // FK to Role table

	AcPermission AcPermissions `gorm:"foreignKey:PermissionID;References:ID"`
	AcPolGrp     AcPolGrp      `gorm:"foreignKey:AcPolGrpID;References:ID"`
	OwnerSpace   Space         `gorm:"foreignKey:OwnerSpaceID;References:SpaceID"`
	OwnerRole    Role          `gorm:"foreignKey:RoleID;References:ID"`
	OwnerTeam    Team          `gorm:"foreignKey:OwnerTeamID;References:TeamID"`
	OwnerUser    User          `gorm:"foreignKey:OwnerUserID;References:UserID"`
}

type Space struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	SpaceID               string         `gorm:"primaryKey;not null"` // NOT NULL	FK to MEMBER table for this space or space unit (type S).
	LegalID               string         // The registered space identifier, given to the space (such as assigned by the government). This may be null for an space unit. This is not the name of the space, which should be stored in the ORGENTITYNAME table.
	Type                  string         `gorm:"not null"`             // NOT NULL P -> personal, B -> business or institution
	Name                  string         `gorm:"unique; not null"`     // Name of the space
	BusinessName          string         `gorm:"unique; default:null"` // The business name of the space. (Unique if value exist)
	Address               string         // Address of the business or institution
	LogoURL               string         // Space logo url
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

type Member struct {
	gorm.Model

	ID         string `gorm:"primaryKey;not null"`
	Type       string `gorm:"size:3"` //  (S-space, U-user, T-teams)
	OptCounter int    `gorm:"size:8"`
}

type Role struct {
	gorm.Model

	ID           string `gorm:"primaryKey;not null"`
	Name         string `gorm:"index:rolename_unique_index,unique"`
	Description  string
	OwnerSpaceID string `gorm:"index:rolename_unique_index,unique"`
	IsOwner      bool
	CreatedBy    string
	UpdatedBy    string
	OptCounter   int `gorm:"size:8"`

	OwnerSpace Space `gorm:"foreignKey:OwnerSpaceID;References:SpaceID"`
}

type Team struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	TeamID      string `gorm:"primaryKey;not null"`                          // PK and FK to MEMBER table (type T).
	OwnerID     string `gorm:"not null; index:teamname_unique_index,unique"` // PK and FK to MEMBER table  (type S).
	Name        string `gorm:"index:teamname_unique_index,unique"`
	Description string
	Update      string
	UpdatedBy   string
	OptCounter  int `gorm:"size:8"`

	Member     Member `gorm:"foreignKey:TeamID;References:ID"`
	OwnerSpace Space  `gorm:"foreignKey:OwnerID;References:SpaceID"`
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

// type Orders struct {
// 	gorm.Model
// 	ID          string `gorm:"primaryKey; not null"`
// 	OwnerUserID string
// 	Status      int // 0 processing  1 sucess 2 failed

// 	User User `gorm:"ForeignKey:OwnerUserID;References:UserID"`
// }

// type OrderItems struct {
// 	gorm.Model
// 	ID                   string `gorm:"primaryKey; not null"`
// 	OrderID              string
// 	BlockID              string
// 	SpaceID              string
// 	InventoryItemID      string
// 	PaymentInfoID        string
// 	TotalPrice           decimal.Decimal ` gorm:"type:numeric"`
// 	VariantBlockID       string
// 	RefundStatus         int // 0 not refundable 1 to be refunded
// 	Status               int // 0 processing  1 sucess 2 failed 3 refunded
// 	SubscriptionID       string
// 	LicenseID            string
// 	InventoryID          string
// 	TransactionDetailsID string
// 	UserPayoutID         string
// 	NetAppblocksShare    decimal.Decimal ` gorm:"type:numeric"`
// 	NetPayoutAmount      decimal.Decimal ` gorm:"type:numeric"`

// 	Order         Orders         `gorm:"ForeignKey:OrderID;References:ID"`
// 	InventoryItem InventoryItems `gorm:"ForeignKey:InventoryItemID;References:ID"`
// 	PaymentInfo   PaymentInfo    `gorm:"ForeignKey:PaymentInfoID;References:ID"`
// 	// Space              Space              `gorm:"ForeignKey:SpaceID;References:ID"`
// 	Block              Block              `gorm:"ForeignKey:BlockID;References:ID"`
// 	VariantBlock       Block              `gorm:"ForeignKey:VariantBlockID;References:ID"`
// 	Inventory          Inventories        `gorm:"ForeignKey:InventoryID;References:ID"`
// 	Subscription       Subscription       `gorm:"ForeignKey:SubscriptionID;References:ID"`
// 	License            License            `gorm:"ForeignKey:LicenseID;References:ID"`
// 	TransactionDetails TransactionDetails `gorm:"ForeignKey:TransactionDetailsID;References:ID"`
// 	UserPayoutsID      UserPayouts        `gorm:"ForeignKey:UserPayoutID;References:ID"`
// }

// type OrderItemDetails struct {
// 	gorm.Model
// 	ID                         string `gorm:"primaryKey; not null"`
// 	InventoryID                string
// 	BlockID                    string
// 	SubscriptionID             string
// 	LicenseID                  string
// 	OrderItemID                string
// 	BlockName                  string
// 	BlockShortName             string
// 	BlockType                  int
// 	ListedPrice                decimal.Decimal ` gorm:"type:numeric"`
// 	SellingPrice               decimal.Decimal ` gorm:"type:numeric"`
// 	DiscountType               int             // 1 flat 2 percent
// 	DiscountValue              decimal.Decimal ` gorm:"type:numeric"`
// 	Currency                   string          // "USD" default
// 	ListingStartDate           time.Time
// 	ListingEndDate             time.Time
// 	LicenseName                string
// 	LicensePeriod              int // in years
// 	SubscriptionName           string
// 	SubscriptionPeriod         int             // in years
// 	InventoryDevelopmentCost   decimal.Decimal ` gorm:"type:numeric"`
// 	InventoryDevelopmentEffort int64

// 	OrderItems   OrderItems   `gorm:"ForeignKey:OrderItemID;References:ID"`
// 	Inventory    Inventories  `gorm:"ForeignKey:InventoryID;References:ID"`
// 	Subscription Subscription `gorm:"ForeignKey:SubscriptionID;References:ID"`
// 	License      License      `gorm:"ForeignKey:LicenseID;References:ID"`
// 	Block        Block        `gorm:"ForeignKey:BlockID;References:ID"`
// }

type AcPermissions struct {
	gorm.Model
	ID           string `gorm:"primaryKey;not null"`
	Description  string `gorm:"size:255; not null;"`
	Name         string
	OptCounter   int `gorm:"size:8"`
	IsPredefined bool
	Type         int // 1 for internal apps and 2 for consumer apps
	DisplayName  string
}

type PerPolGrps struct {
	gorm.Model
	ID             string `gorm:"primaryKey;not null"`
	AcPermissionID string
	AcPolGrpID     string

	AcPolGrp     AcPolGrp      `gorm:"foreignKey:AcPolGrpID;References:ID"`
	AcPermission AcPermissions `gorm:"foreignKey:AcPermissionID;References:ID"`
}

// type PaymentInfo struct {
// 	gorm.Model
// 	ID                 string `gorm:"primaryKey; not null"`
// 	PaymentGatewayID   string `gorm:"not null; UniqueIndex:payment_gateway_id"`
// 	PaymentGatewayType int
// }

// type ProfitSharingDetails struct {
// 	gorm.Model
// 	ID                   string          `gorm:"primaryKey; not null"`
// 	AppblocksPercentage  decimal.Decimal ` gorm:"type:numeric"`
// 	AppblocksFixedAmount decimal.Decimal ` gorm:"type:numeric"`
// 	StripePercentage     decimal.Decimal ` gorm:"type:numeric"`
// 	StripeFixedAmount    decimal.Decimal ` gorm:"type:numeric"`
// 	TotalPercentage      decimal.Decimal ` gorm:"type:numeric"`
// 	TotalFixedAmount     decimal.Decimal ` gorm:"type:numeric"`
// 	Status               int             //1 for active 2 for inactive
// 	CreatedBy            string
// 	UpdatedBy            string
// }

// type License struct {
// 	gorm.Model
// 	ID          string `gorm:"primaryKey; not null"`
// 	Name        string
// 	Period      int // in years
// 	Description string
// }

// type Subscription struct {
// 	gorm.Model
// 	ID          string `gorm:"primaryKey; not null"`
// 	Name        string
// 	Period      int // in years
// 	Description string
// 	Type        int // 1 for update subscription

// }

// type Inventories struct {
// 	gorm.Model
// 	ID                string `gorm:"primaryKey; not null"`
// 	BlockID           string
// 	Status            int // 1 listed 2 unlisted
// 	ItemName          string
// 	ListingStartDate  time.Time
// 	ListingEndDate    time.Time
// 	DevelopmentCost   decimal.Decimal ` gorm:"type:numeric"`
// 	DevelopmentEffort int64
// 	CreatedBy         string

// 	User  User  `gorm:"ForeignKey:CreatedBy;References:UserID"`
// 	Block Block `gorm:"ForeignKey:BlockID;References:ID"`
// }

// type InventoryItems struct {
// 	gorm.Model
// 	ID               string `gorm:"primaryKey; not null"`
// 	InventoryID      string
// 	SubscriptionID   string
// 	LicenseID        string
// 	Status           int // 1 listed 2 unlisted
// 	ItemName         string
// 	ListedPrice      decimal.Decimal ` gorm:"type:numeric"`
// 	SellingPrice     decimal.Decimal ` gorm:"type:numeric"`
// 	DiscountType     int             // 1 flat 2 percent
// 	DiscountValue    decimal.Decimal ` gorm:"type:numeric"`
// 	Currency         string          // "USD" default
// 	ListingStartDate time.Time
// 	ListingEndDate   time.Time
// 	CreatedBy        string

// 	User         User         `gorm:"ForeignKey:CreatedBy;References:UserID"`
// 	Inventory    Inventories  `gorm:"ForeignKey:InventoryID;References:ID"`
// 	Subscription Subscription `gorm:"ForeignKey:SubscriptionID;References:ID"`
// 	License      License      `gorm:"ForeignKey:LicenseID;References:ID"`
// }

// // Model for Apps Table
// type App struct {
// 	gorm.Model
// 	ID               string `gorm:"primaryKey; not null"`
// 	Name             string
// 	DisplayName      string
// 	AppID            string `gorm:"unique; not null"`
// 	Status           int
// 	DeploymentMode   int // 0 saas 1 business 2 both
// 	AppBlockID       string
// 	Language         string
// 	PricingType      int // 0 free 1 paid
// 	PrivacyPolicyURL string
// 	CompanyName      string
// 	ContactEmail     string
// 	Block            Block    `gorm:"ForeignKey:AppBlockID;References:ID"`
// 	Entity           Entities `gorm:"foreignKey:ID;References:EntityID;unique;not null"`
// }

// type BlockUpdatesPull struct {
// 	gorm.Model
// 	ID             string `gorm:"primaryKey"`
// 	AppID          string `gorm:"not null;uniqueIndex:bup_app_id_block_id_block_version_id"`
// 	BlockID        string `gorm:"not null;uniqueIndex:bup_app_id_block_id_block_version_id"`
// 	BlockVersionID string `gorm:"not null;uniqueIndex:bup_app_id_block_id_block_version_id"`

// 	Block        Block         `gorm:"ForeignKey:BlockID;References:ID"`
// 	BlockVersion BlockVersions `gorm:"ForeignKey:BlockVersionID;References:ID"`
// 	App          App           `gorm:"ForeignKey:AppID;References:ID"`
// }

// type Carts struct {
// 	gorm.Model
// 	ID        string `gorm:"primaryKey; not null"`
// 	CreatedBy string `gorm:"not null; UniqueIndex:user_id"`

// 	User User `gorm:"ForeignKey:CreatedBy;References:UserID"`
// }

// type CartItems struct {
// 	gorm.Model
// 	ID              string `gorm:"primaryKey; not null"`
// 	CartID          stringCartItems
// 	InventoryItemID string
// 	OwnerSpaceID    string

// 	InventoryItem InventoryItems `gorm:"ForeignKey:InventoryItemID;References:ID"`
// 	Carts         Carts          `gorm:"ForeignKey:CartID;References:ID"`
// 	Space         Space          `gorm:"ForeignKey:OwnerSpaceID;References:SpaceID"`
// }

// type CartItemHierarchies struct {
// 	gorm.Model
// 	ID                string `gorm:"primaryKey; not null"`
// 	ParentID          string
// 	ChildID           string
// 	ParentOrderItemID string

// 	ParentCartItem CartItems `gorm:"ForeignKey:ParentID;References:ID"`
// 	ChildCartItem  CartItems `gorm:"ForeignKey:ChildID;References:ID"`
// 	// OrderItem      OrderItems `gorm:"ForeignKey:ParentOrderItemID;References:ID"`
// }

// type TransactionDetails struct {
// 	gorm.Model
// 	ID                       string `gorm:"primaryKey; not null"`
// 	StripeTransactionID      string
// 	Status                   int             // 0 not credited 1 credited in stripe balance
// 	StripeFeesAmount         decimal.Decimal ` gorm:"type:numeric"`
// 	CustomerAmountPaid       decimal.Decimal ` gorm:"type:numeric"`
// 	StripeNetAmountReceived  decimal.Decimal ` gorm:"type:numeric"`
// 	AppBlocksProfitDetailsID string

// 	AppblocksProfitDetails AppblocksProfitDetails `gorm:"ForeignKey:AppBlocksProfitDetailsID;References:ID"`
// }

// type UserPayouts struct {
// 	gorm.Model
// 	ID                     string `gorm:"primaryKey; not null"`
// 	OwnerUserID            string
// 	UserConnectedAccountID string
// 	OwnerSpaceID           string
// 	StripeFeesAmount       decimal.Decimal ` gorm:"type:numeric"`
// 	PayoutAmount           decimal.Decimal ` gorm:"type:numeric"`

// 	Status int // payout created 2// payout paid 3// payout failed

// 	OwnerUser  User  `gorm:"ForeignKey:OwnerUserID;References:UserID"`
// 	OwnerSpace Space `gorm:"foreignKey:OwnerSpaceID;References:SpaceID"`
// }

// type AppblocksProfitDetails struct {
// 	gorm.Model
// 	ID           string          `gorm:"primaryKey; not null"`
// 	ProfitAmount decimal.Decimal ` gorm:"type:numeric"`
// }

// type OrderItemHierarchies struct {
// 	gorm.Model
// 	ID       string `gorm:"primaryKey; not null"`
// 	ParentID string
// 	ChildID  string

// 	ParentOrderItem OrderItems `gorm:"ForeignKey:ParentID;References:ID"`
// 	ChildOrderItem  OrderItems `gorm:"ForeignKey:ChildID;References:ID"`
// }

// shield models

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
	DeletedAt    time.Time
	OwnerSpaceID string
	ID           int

	OwnerSpace Space `gorm:"foreignKey:OwnerSpaceID;References:SpaceID"`
}

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

// Model for AppPermissions Table
type AppPermission struct {
	AppId        string `gorm:"primary_key; size:255; not null"`
	PermissionId string `gorm:"primary_key; size:255; not null"`
	Mandatory    bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Model for AppUserPermissions Table
type AppUserPermission struct {
	UserId       string `gorm:"primary_key; size:255; not null"`
	AppId        string `gorm:"primary_key; size:255; not null"`
	PermissionId string `gorm:"primary_key; size:255; not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// type StripeDetails struct {
// 	ID                 string `gorm:"primaryKey; size:255"` // FK to MEMBER table (type U).
// 	OwnerUserID        string `gorm:"size:255; not null; unique"`
// 	StripeCustomerID   string `gorm:"size:255; not null; unique"`
// 	ConnectedAccountID string `gorm:"size:255"`
// 	CreatedAt          time.Time
// 	UpdatedAt          time.Time

// 	User User `gorm:"foreignKey:OwnerUserID;References:UserID"`
// }

// Model for UserProvider Table
type UserProvider struct {
	UserId    string `gorm:"primary_key; size:255; not null"`
	Provider  int    `gorm:"primary_key; not null; default:1"` // 1 - shield, 2 - google
	CreatedAt time.Time
	UpdatedAt time.Time
}

type MemberRole struct {
	gorm.Model

	ID           string `gorm:"primaryKey;not null"`
	OwnerUserID  string `gorm:"index:member_role_unique_index,unique"` // PK and FK to MEMBER table (type U).
	RoleID       string `gorm:"index:member_role_unique_index,unique"` // NOT NULL	FK to Role.
	OwnerSpaceID string `gorm:"index:member_role_unique_index,unique"` // PK and FK to MEMBER table (type U).
	OptCounter   int    `gorm:"size:8"`

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

type TeamMember struct {
	gorm.Model
	ID          string `gorm:"primaryKey;not null"`
	OwnerTeamID string `gorm:"index:team_member_unique_index,unique"` // PK and FK to Team.
	MemberID    string `gorm:"index:team_member_unique_index,unique"` // PK and FK to MEMBER table (type U).
	IsOwner     bool   `gorm:"default:false"`
	OptCounter  int    `gorm:"size:8"`

	Team       Team `gorm:"foreignKey:OwnerTeamID;References:TeamID"`
	UserMember User `gorm:"foreignKey:MemberID;References:UserID"`
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

type AcResource struct {
	gorm.Model
	ID              string `gorm:"primaryKey;not null"`
	Name            string
	Description     string
	Path            string
	FunctionName    string
	EntityName      string
	FunctionMethod  string
	Version         string
	OptCounter      int `gorm:"size:8"`
	OwnerAppID      string
	IsAuthorised    int
	IsAuthenticated int
	OwnerApp        ShieldApp `gorm:"foreignKey:OwnerAppID;References:AppId"`
}

type AcResGrp struct {
	gorm.Model
	ID           string `gorm:"primaryKey;not null"`
	OwnerSpaceID string // FK to spaces table
	Name         string
	Description  string
	IsPredefined bool
	OptCounter   int `gorm:"size:8"`
	Type         int // 1 for internal apps and 2 for consumer apps

	OwnerSpace Space `gorm:"foreignKey:OwnerSpaceID;References:SpaceID"`
}

type AcResGpRes struct {
	gorm.Model
	ID           string `gorm:"primaryKey;not null"`
	AcResGrpID   string
	AcResourceID string // FK to MEMBER table
	OptCounter   int    `gorm:"size:8"`

	AcResource AcResource `gorm:"foreignKey:AcResourceID;references:ID"`
	AcResGrp   AcResGrp   `gorm:"foreignKey:AcResGrpID;references:ID"`
}

type AcResAction struct {
	gorm.Model
	ID           string `gorm:"primaryKey;not null"`
	AcActionID   string
	AcResourceID string // FK to MEMBER table
	OptCounter   int    `gorm:"size:8"`

	AcResource AcResource `gorm:"foreignKey:AcResourceID;References:ID"`
	AcAction   AcAction   `gorm:"foreignKey:AcActionID;References:ID"`
}

type AcAction struct {
	gorm.Model
	ID          string `gorm:"primaryKey;not null"`
	Name        string
	Description string
	OptCounter  int `gorm:"size:8"`
	OwnerAppID  string
	OwnerApp    ShieldApp `gorm:"foreignKey:OwnerAppID;References:AppId"`
}

type AcActGrp struct {
	gorm.Model
	ID           string `gorm:"primaryKey;not null"`
	OwnerSpaceID string // FK to spaces table
	Description  string
	Name         string
	IsPredefined bool
	OptCounter   int `gorm:"size:8"`
	Type         int // 1 for internal apps and 2 for consumer apps

	OwnerSpace Space `gorm:"foreignKey:OwnerSpaceID;References:SpaceID"`
}

type ActGpAction struct {
	gorm.Model
	ID         string `gorm:"primaryKey"`
	AcActGrpID string
	AcActionID string
	OptCounter int `gorm:"size:8"`

	AcAction AcAction `gorm:"foreignKey:AcActionID;References:ID"`
	AcActGrp AcActGrp `gorm:"foreignKey:AcActGrpID;References:ID"`
}

type AcPolicy struct {
	gorm.Model
	ID           string `gorm:"primaryKey;not null"`
	AcActGrpID   string
	AcResGrpID   string
	OwnerSpaceID string // FK to MEMBER table
	CreatedBy    string // FK to USERS TABLE
	UpdatedBy    string // FK TO USERS TABLE
	Name         string
	Description  string
	Path         string
	OptCounter   int `gorm:"size:8"`
	IsPredefined bool
	Type         int // 1 for internal apps and 2 for consumer apps

	AcActionGroup   AcActGrp `gorm:"foreignKey:AcActGrpID;References:ID"`
	AcResourceGroup AcResGrp `gorm:"foreignKey:AcResGrpID;References:ID"`
	OwnerSpace      Space    `gorm:"foreignKey:OwnerSpaceID;References:SpaceID"`
	CreatedUser     User     `gorm:"foreignKey:CreatedBy;References:UserID"`
	UpdatedUser     User     `gorm:"foreignKey:UpdatedBy;References:UserID"`
}
type AcPolGrp struct {
	gorm.Model
	ID           string `gorm:"primaryKey;not null"`
	OwnerSpaceID string // FK to MEMBER table
	Description  string `gorm:"size:255; not null;"`
	Name         string
	OptCounter   int `gorm:"size:8"`
	IsPredefined bool
	Type         int // 1 for internal apps and 2 for consumer apps
	EntityType   int // 0 for non entity based policies 1 for block, 2 for app, 3 for environment
	DisplayName  string
	EntityTypes  pq.Int64Array `gorm:"type:integer[]"`
	// type changed to array and renamed to types  0 for non entity based policies 1 for block, 2 for app, 3 for environment

	OwnerSpace Space `gorm:"foreignKey:OwnerSpaceID;References:SpaceID"`
}

type Entities struct {
	EntityID  string `gorm:"primaryKey;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	Type      int64          // 1 for block, 2 for app, 3 for environment
}

// type PolGrpEntityMapping struct {
// 	gorm.Model

// 	ID            string `gorm:"primaryKey;not null"`
// 	OwnerEntityID string
// 	PolGrpID      string

// 	PolicyGroup AcPolGrp `gorm:"foreignKey:PolGrpID;References:ID"`
// 	Entities    Entities `gorm:"foreignKey:OwnerEntityID;References:EntityID"`
// }

// type Environment struct {
// 	gorm.Model
// 	ID              string `gorm:"primaryKey"`
// 	Name            string
// 	AppID           string
// 	PredefinedEnvID string

// 	App           App           `gorm:"ForeignKey:AppID;References:ID"`
// 	PredefinedEnv PredefinedEnv `gorm:"ForeignKey:PredefinedEnvID;References:ID"`
// 	Entity        Entities      `gorm:"foreignKey:ID;References:EntityID;unique;not null"`
// }

type PredefinedEnv struct {
	gorm.Model
	ID   string `gorm:"primaryKey; not null"`
	Name string
}

type PolGpPolicy struct {
	gorm.Model
	ID         string `gorm:"primaryKey;not null"`
	AcPolicyID string
	AcPolGrpID string // FK to MEMBER table
	OptCounter int    `gorm:"size:8"`

	AcPolGrp AcPolGrp `gorm:"foreignKey:AcPolGrpID;References:ID"`
	AcPolicy AcPolicy `gorm:"foreignKey:AcPolicyID;References:ID"`
}

type Invites struct {
	gorm.Model
	ID         string `gorm:"primaryKey;not null"`
	Notes      string
	CreatedBy  string
	Status     int // 1-pending 2 complete 3 declined
	ExpiresAt  time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
	InviteType int //1-email invyt 2 invite link
	InviteLink string
	InviteCode string
	Email      string

	CreatedUser User `gorm:"foreignKey:CreatedBy;References:UserID"`
}

type InviteDetails struct {
	gorm.Model
	ID             string `gorm:"primaryKey;not null"`
	InviteID       string
	InvitedSpaceID string
	InvitedTeamID  string `gorm:"default: null"`
	InvitedRoleID  string `gorm:"default: null"`
	Email          string
	CreatedAt      time.Time
	UpdatedAt      time.Time

	Invite Invites `gorm:"foreignKey:InviteID;References:ID"`
	Space  Space   `gorm:"foreignKey:InvitedSpaceID;References:SpaceID"`
	Team   Team    `gorm:"foreignKey:InvitedTeamID;References:TeamID"`
	Role   Role    `gorm:"foreignKey:InvitedRoleID;References:ID"`
}

// type InventoryApprovals struct {
// 	gorm.Model
// 	ID              string `gorm:"primaryKey;not null"`
// 	Type            int    // free: 0, paid 1
// 	Status          int    // approved: 1, pending: 2, rejected: 3
// 	RejectionType   int    // version release: 1, version release + meta data changes: 2,  meta data changes: 3
// 	RejectionReason string
// 	MetaData        pgtype.JSONB `gorm:"type:jsonb"`
// 	UpdateType      int          // 1: metadata, 2: new release version added, 3: first release
// 	BlockVersionID  string

// 	Version BlockVersions `gorm:"ForeignKey:BlockVersionID;References:ID"`
// }

type AppSpaceMapping struct {
	gorm.Model
	ID      string `gorm:"primaryKey; not null"`
	SpaceID string
	AppID   string
	// App     App `gorm:"ForeignKey:AppID;References:ID"`
}

// type AppCategoryMapping struct {
// 	gorm.Model
// 	ID         string `gorm:"primaryKey; not null"`
// 	CategoryID string
// 	AppID      string

// 	// App         App         `gorm:"ForeignKey:AppID;References:ID"`
// 	AppCategory AppCategory `gorm:"ForeignKey:CategoryID;References:ID"`
// }

// type AppCategory struct {
// 	gorm.Model
// 	ID           string `gorm:"primaryKey; not null"`
// 	CategoryName string
// 	Description  string
// }

// type AppCategoryHierarchy struct {
// 	gorm.Model
// 	ID                 string `gorm:"primaryKey; not null"`
// 	Ancestor           string
// 	Descendant         string
// 	NumLevel           int
// 	AncestorCategory   AppCategory `gorm:"ForeignKey:Ancestor;References:ID"`
// 	DescendantCategory AppCategory `gorm:"ForeignKey:Descendant;References:ID"`
// }

// type EnvVariables struct {
// 	gorm.Model
// 	ID            string `gorm:"primaryKey"`
// 	Name          string
// 	Value         string
// 	EnvironmentID string
// 	Status        bool
// 	Environment   Environment `gorm:"ForeignKey:EnvironmentID;References:ID"`
// }

// type VMInfo struct {
// 	gorm.Model
// 	ID                 string `gorm:"primaryKey"`
// 	AppID              string
// 	PublicIPAddress    string
// 	PublicDNSName      string
// 	PublicDNSUser      string
// 	PemPrivateKeyValue string
// 	InstanceID         string
// 	Status             int64 // 0 : pending	16 : running 32 : shutting-down 48 : terminated 64 : stopping 80 : stopped
// 	InstanceData       datatypes.JSON
// 	App                App `gorm:"ForeignKey:AppID;References:ID"`
// }

// type VMUserInfo struct {
// 	gorm.Model
// 	ID                      string `gorm:"primaryKey"`
// 	Password                string
// 	UserName                string
// 	EnvironmentID           string
// 	VMInfoID                string
// 	CurrentlyDeployedFolder int         // 1-green, 2-blue
// 	Environment             Environment `gorm:"ForeignKey:EnvironmentID;References:ID"`
// 	VMInfo                  VMInfo      `gorm:"ForeignKey:VMInfoID;References:ID"`
// }

// type VMPortInfo struct {
// 	gorm.Model
// 	ID           string `gorm:"primaryKey"`
// 	Port         int
// 	VMUserInfoID string
// 	VMUserInfo   VMUserInfo `gorm:"ForeignKey:VMUserInfoID;References:ID"`
// 	Status       int        // 1 active 2 deleted
// }

// type DNSRecordInfo struct {
// 	gorm.Model
// 	ID            string `gorm:"primaryKey"`
// 	Domain        string
// 	Content       string
// 	DNSRecordID   string
// 	AppID         string
// 	EnvironmentID string
// 	DNSRecord     datatypes.JSON
// 	App           App         `gorm:"ForeignKey:AppID;References:ID"`
// 	Environment   Environment `gorm:"ForeignKey:EnvironmentID;References:ID"`
// }

// type DomainInfo struct {
// 	gorm.Model
// 	ID            string `gorm:"primaryKey"`
// 	SubDomain     string
// 	URL           string
// 	EnvironmentID string
// 	IsDefault     bool
// 	Type          int         // 0 frontend 1 backend
// 	Environment   Environment `gorm:"ForeignKey:EnvironmentID;References:ID"`
// }

// type DatabaseInstanceInfo struct {
// 	gorm.Model
// 	ID                   string `gorm:"primaryKey"`
// 	Name                 string
// 	AppID                string
// 	AwsARN               string
// 	DBInstanceIdentifier string
// 	DBInstanceResourceID string
// 	Status               int // 1 active 2 deleted
// 	AwsEndPoint          string
// 	DNSRecord            datatypes.JSON
// 	DNSRecordID          string
// 	Domain               string
// 	Memory               int
// 	CoreCount            int
// 	Vcpu                 int
// 	App                  App `gorm:"ForeignKey:AppID;References:ID"`
// }

// type DatabaseInfo struct {
// 	gorm.Model
// 	ID            string `gorm:"primaryKey"`
// 	Name          string
// 	EnvironmentID string
// 	Status        bool
// 	Environment   Environment `gorm:"ForeignKey:EnvironmentID;References:ID"`
// }

// type AwsRdsConfig struct {
// 	gorm.Model
// 	ID                 string `gorm:"primaryKey"`
// 	InstanceClass      string
// 	CoreCount          int
// 	Vcpu               int
// 	CPUCreditsPerHour  int
// 	Memory             int
// 	NetworkPerformance string
// 	Engine             string // currently 'postgres' only supported
// 	EngineVersion      string // possible engine version for the instance class
// 	DefaultPort        int    // default port for the given db engine
// 	DefaultStorage     int    // default storage allocation for this instance class
// }

// type DeployLog struct {
// 	gorm.Model
// 	ID            string `gorm:"primaryKey"`
// 	Tags          string
// 	EnvironmentID string
// 	AppVersion    string
// 	Status        int // 0 - inactive 1 -currently released and active
// 	IPAddress     string
// 	IPRoute       string
// 	IPPort        string
// 	MetaData      datatypes.JSON
// 	Environment   Environment `gorm:"ForeignKey:EnvironmentID;References:ID"`
// }

// type DeployHistory struct {
// 	gorm.Model
// 	ID           string `gorm:"primaryKey"`
// 	ReleaseNotes string
// 	Tags         string
// 	DeployLogID  string
// 	UserName     string
// 	UserID       string
// 	DeployLog    DeployLog `gorm:"ForeignKey:DeployLogID;References:ID"`
// }

// type UploadBlocks struct {
// 	gorm.Model
// 	ID           string `gorm:"primaryKey"`
// 	DeployLogID  string
// 	BlockVersion string
// 	BlockID      string
// 	ObjectKey    string
// 	IPAddress    string
// 	IPRoute      string
// 	IPPort       string
// 	BlockType    int // app - 1 , ui-container - 2,ui-elements 3  fn - 4, data - 5,function shared block -6
// 	MetaData     datatypes.JSON
// 	Block        Block     `gorm:"ForeignKey:BlockID;References:ID"`
// 	DeployLog    DeployLog `gorm:"ForeignKey:DeployLogID;References:ID"`
// 	Status       int       // 1-healthy and live , 0 inactive
// 	URL          string
// }

// type AppRelease struct {
// 	gorm.Model

// 	ID          string `gorm:"primaryKey; not null"`
// 	AppID       string `gorm:"not null"`
// 	DeployLogID string // release candidate ID, link to deploy log and all uploaded block details
// 	ReleaseNote string `gorm:"not null"`
// 	Version     string `gorm:"not null"`
// 	Status      int    // 0 - pending 1 - active 2 - inactive

// 	App       App       `gorm:"ForeignKey:AppID;References:ID"`
// 	DeployLog DeployLog `gorm:"ForeignKey:DeployLogID;References:ID"`
// }

// type AppReleaseMedia struct {
// 	gorm.Model

// 	ID           string `gorm:"primaryKey; not null"`
// 	AppReleaseID string `gorm:"not null"`
// 	Order        int
// 	DeviceType   int    // 0 - desktop, 1- tablet, 2 - mobile
// 	Key          string // s3 object key
// 	MediaType    int    // 0 - image, 1 - video

// 	AppRelease AppRelease `gorm:"ForeignKey:AppReleaseID;References:ID"`
// }

// type SubmitLog struct {
// 	gorm.Model

// 	ID           string `gorm:"primaryKey; not null"`
// 	SubmittedBy  string `gorm:"not null"`
// 	ReviewedBy   string
// 	SubmitType   int // 0 - release 1 - app 2 - logo
// 	ReviewStatus int // 0 - pending 1 - approved 2 - rejected
// }

// type SubmitLogMetaData struct {
// 	gorm.Model

// 	ID           string `gorm:"primaryKey; not null"`
// 	SubmitLogID  string `gorm:"not null"`
// 	AppID        string `gorm:"default:null"`
// 	AppReleaseID string `gorm:"default:null"`
// 	DeployLogID  string `gorm:"default:null"`

// 	App        App        `gorm:"ForeignKey:AppID;References:ID"`
// 	AppRelease AppRelease `gorm:"ForeignKey:AppReleaseID;References:ID"`
// 	DeployLog  DeployLog  `gorm:"ForeignKey:DeployLogID;References:ID"`
// 	SubmitLog  SubmitLog  `gorm:"ForeignKey:SubmitLogID;References:ID"`
// }

// type SubmitLogPayload struct {
// 	gorm.Model

// 	ID            string `gorm:"primaryKey; not null"`
// 	SubmitLogID   string `gorm:"not null"`
// 	Comment       string
// 	AttachmentURL string
// 	PayloadType   int // 0 - submit 1 -review

// 	SubmitLog SubmitLog `gorm:"ForeignKey:SubmitLogID;References:ID"`
// }

// type PendingUpdate struct {
// 	gorm.Model

// 	ID           string `gorm:"primaryKey; not null"`
// 	AppID        string `gorm:"default:null"`
// 	AppReleaseID string `gorm:"default:null"`
// 	UpdateData   datatypes.JSON
// 	Type         int // 0 - release 1 - app 2 - logo

// 	App        App        `gorm:"ForeignKey:AppID;References:ID"`
// 	AppRelease AppRelease `gorm:"ForeignKey:AppReleaseID;References:ID"`
// }
// type CADDetails struct {
// 	gorm.Model

// 	ID            string `gorm:"primaryKey; not null"`
// 	AppID         string `gorm:"not null"`
// 	SpaceID       string `gorm:"not null"`
// 	EnvironmentID string `gorm:"not null; uniqueIndex:cad_environment_id_container_name"`
// 	ContainerName string `gorm:"not null; uniqueIndex:cad_environment_id_container_name"`

// 	RepositoryID   string
// 	RepositoryName string
// 	RepositoryURI  string
// 	RepositoryData datatypes.JSON

// 	ClusterARN  string
// 	ClusterName string
// 	ClusterData datatypes.JSON

// 	TaskDefinitionARN    string
// 	TaskDefinitionFamily string
// 	TaskDefinitionData   datatypes.JSON

// 	ServiceARN  string
// 	ServiceName string
// 	ServiceData datatypes.JSON

// 	LoadBalancerARN  string
// 	LoadBalancerName string
// 	LoadBalancerData datatypes.JSON

// 	TargetGroupARN  string
// 	TargetGroupName string
// 	TargetGroupData datatypes.JSON

// 	AutoScalingPolicies          datatypes.JSON
// 	AutoScalableTargetResourceID string

// 	App         App         `gorm:"ForeignKey:AppID;References:ID"`
// 	Environment Environment `gorm:"ForeignKey:EnvironmentID;References:ID"`
// }

// type PreviewEnvVariables struct {
// 	gorm.Model
// 	ID             string        `gorm:"primaryKey"`
// 	Name           string        // env name
// 	Value          string        // env value
// 	BlockVersionID string        `gorm:"not null"`
// 	BlockVersion   BlockVersions `gorm:"ForeignKey:BlockVersionID;References:ID"`
// }

// type StartPreviewQueue struct {
// 	gorm.Model
// 	ID        string       `gorm:"primaryKey"`
// 	Payload   pgtype.JSONB `gorm:"type:jsonb"`
// 	CreatedAt time.Time
// 	UpdatedAt time.Time
// }

// type InvalidQueueItems struct {
// 	gorm.Model
// 	ID        string       `gorm:"primaryKey"`
// 	Payload   pgtype.JSONB `gorm:"type:jsonb"`
// 	CreatedAt time.Time
// 	UpdatedAt time.Time
// }

// type BlockPreviewStatus struct {
// 	BlockVersionID string `gorm:"primaryKey"`
// 	CreatedAt      time.Time
// 	UpdatedAt      time.Time
// 	BlockVersion   BlockVersions `gorm:"ForeignKey:BlockVersionID;References:ID"`
// 	Status         int           // 1 means preview is in build phase.2 means preview is running
// 	Metadata       pgtype.JSONB  `gorm:"type:jsonb"`
// }

// type PreviewStartLogs struct {
// 	gorm.Model
// 	ID            string       `gorm:"primaryKey"`
// 	Payload       pgtype.JSONB `gorm:"type:jsonb"`
// 	CreatedBy     string       `gorm:"not null"`
// 	CreatedByUser User         `gorm:"ForeignKey:CreatedBy;References:UserID"`
// }