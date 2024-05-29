package main

import (
	"log"

	"github.com/appblocks-hub/SHIELD/functions/pghandling"
	"github.com/appblocks-hub/SHIELD/models"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

func main() {
	// Load env vars
	err := godotenv.Load("./.env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	db, err := pghandling.SetupDB()
	if err != nil {
		log.Fatalf("Error init db: %v", err)
	}
	Migrate(db)
}

func Migrate(db *gorm.DB) {
	// dropping tables
	// db.Migrator().DropTable(&general.User{})
	// db.Migrator().DropTable(&models.Entities{})
	// db.Migrator().DropTable(&models.App{})

	// db.Migrator().DropTable(&general.ShieldApp{})
	// db.Migrator().DropTable(&general.Permission{})
	// db.Migrator().DropTable(&general.AppPermission{})
	// db.Migrator().DropTable(&general.AppUserPermission{})

	//db.AutoMigrate(&models.User{})
	// db.AutoMigrate(&models.ShieldApp{})
	//db.AutoMigrate(&models.Permission{})
	//db.AutoMigrate(&models.AppPermission{})
	//db.AutoMigrate(&models.AppUserPermission{})
	//db.AutoMigrate(&models.UserProvider{})

	//db.AutoMigrate(&models.Member{})
	//db.AutoMigrate(&models.Space{})
	//db.AutoMigrate(&models.Role{})
	//db.AutoMigrate(&models.MemberRole{})
	//db.AutoMigrate(&models.SpaceMember{})
	//db.AutoMigrate(&models.DefaultUserSpace{})
	//db.AutoMigrate(&models.Team{})
	//db.AutoMigrate(&models.TeamMember{})
	//db.AutoMigrate(&models.AcResource{})
	//db.AutoMigrate(&models.AcResGrp{})
	//db.AutoMigrate(&models.AcResGpRes{})
	//db.AutoMigrate(&models.AcResAction{})
	//db.AutoMigrate(&models.AcAction{})
	//db.AutoMigrate(&models.AcActGrp{})
	//db.AutoMigrate(&models.ActGpAction{})
	//db.AutoMigrate(&models.AcPolicy{})
	// db.AutoMigrate(&models.AcPolGrp{})
	// db.AutoMigrate(&models.PolGpPolicy{})

	polGrpPolErr := db.AutoMigrate(&models.PolGpPolicy{})
	if polGrpPolErr != nil {
		log.Fatalf("Error  AutoMigrate PolGpPolicy %v", polGrpPolErr)
	}

	acPolGrpErr := db.AutoMigrate(&models.AcPolGrp{})
	if acPolGrpErr != nil {
		log.Fatalf("Error  AutoMigrate AcPolGrp %v", acPolGrpErr)
	}

	acPolErr := db.AutoMigrate(&models.AcPolicy{})
	if acPolErr != nil {
		log.Fatalf("Error  AutoMigrate AcPolicy %v", acPolErr)
	}

	acGpAcErr := db.AutoMigrate(&models.ActGpAction{})
	if acGpAcErr != nil {
		log.Fatalf("Error  AutoMigrate ActGpAction %v", acGpAcErr)
	}

	acAcGrpErr := db.AutoMigrate(&models.AcActGrp{})
	if acAcGrpErr != nil {
		log.Fatalf("Error  AutoMigrate AcActGrp %v", acAcGrpErr)
	}

	acAcErr := db.AutoMigrate(&models.AcAction{})
	if acAcErr != nil {
		log.Fatalf("Error  AutoMigrate AcAction %v", acAcErr)
	}

	acReAcErr := db.AutoMigrate(&models.AcResAction{})
	if acReAcErr != nil {
		log.Fatalf("Error  AutoMigrate AcResAction %v", acReAcErr)
	}

	acReGrpResErr := db.AutoMigrate(&models.AcResGpRes{})
	if acReGrpResErr != nil {
		log.Fatalf("Error  AutoMigrate AcResGpRes %v", acReGrpResErr)
	}

	acReGrpErr := db.AutoMigrate(&models.AcResGrp{})
	if acReGrpErr != nil {
		log.Fatalf("Error  AutoMigrate AcResGrp %v", acReGrpErr)
	}

	acRErr := db.AutoMigrate(&models.AcResource{})
	if acRErr != nil {
		log.Fatalf("Error  AutoMigrate AcResource %v", acRErr)
	}

	tmMbErr := db.AutoMigrate(&models.TeamMember{})
	if tmMbErr != nil {
		log.Fatalf("Error  AutoMigrate TeamMember %v", tmMbErr)
	}

	teamErr := db.AutoMigrate(&models.Team{})
	if teamErr != nil {
		log.Fatalf("Error  AutoMigrate Team %v", teamErr)
	}

	deUsrSpErr := db.AutoMigrate(&models.DefaultUserSpace{})
	if deUsrSpErr != nil {
		log.Fatalf("Error  AutoMigrate DefaultUserSpace %v", deUsrSpErr)
	}

	mbrRErr := db.AutoMigrate(&models.MemberRole{})
	if mbrRErr != nil {
		log.Fatalf("Error  AutoMigrate MemberRole  %v", mbrRErr)
	}

	rErr := db.AutoMigrate(&models.Role{})
	if rErr != nil {
		log.Fatalf("Error AutoMigrate Role %v", rErr)
	}

	spErr := db.AutoMigrate(&models.Space{})
	if spErr != nil {
		log.Fatalf("Error AutoMigrate Space %v", spErr)
	}

	mbrErr := db.AutoMigrate(&models.Member{})
	if mbrErr != nil {
		log.Fatalf("Error AutoMigrate Member %v", mbrErr)
	}

	usrPrErr := db.AutoMigrate(&models.UserProvider{})
	if usrPrErr != nil {
		log.Fatalf("Error AutoMigrate UserProvider %v", usrPrErr)
	}

	appUsrPerErr := db.AutoMigrate(&models.AppUserPermission{})
	if appUsrPerErr != nil {
		log.Fatalf("Error AutoMigrate AppUserPermission %v", appUsrPerErr)
	}

	appPerErr := db.AutoMigrate(&models.AppPermission{})
	if appPerErr != nil {
		log.Fatalf("Error AutoMigrate AppPermission %v", appPerErr)
	}

	perErr := db.AutoMigrate(&models.Permission{})
	if perErr != nil {
		log.Fatalf("Error AutoMigrate Permission %v", perErr)
	}

	shldErr := db.AutoMigrate(&models.ShieldApp{})
	if shldErr != nil {
		log.Fatalf("Error AutoMigrate ShieldApp %v", shldErr)
	}
	usrErr := db.AutoMigrate(&models.User{})
	if usrErr != nil {
		log.Fatalf("Error AutoMigrate User %v", usrErr)
	}

	acPerErr := db.AutoMigrate(&models.AcPermissions{})
	if acPerErr != nil {
		log.Fatalf("Error AutoMigrate AcPermissions %v", acPerErr)
	}
	perPolGrpErr := db.AutoMigrate(&models.PerPolGrps{})
	if perPolGrpErr != nil {
		log.Fatalf("Error AutoMigrate PerPolGrps %v", perPolGrpErr)
	}

	polGrpSubErr := db.AutoMigrate(&models.AcPolGrpSub{})
	if polGrpSubErr != nil {
		log.Fatalf("Error AutoMigrate AcPolGrpSub %v", polGrpSubErr)
	}

	InviteDetailsErr := db.AutoMigrate(&models.InviteDetails{})
	if InviteDetailsErr != nil {
		log.Fatalf("Error AutoMigrate InviteDetails %v", InviteDetailsErr)
	}

	InviteErr := db.AutoMigrate(&models.Invites{})
	if InviteErr != nil {
		log.Fatalf("Error AutoMigrate Invites %v", InviteErr)
	}

	SpaceMemberErr := db.AutoMigrate(&models.SpaceMember{})
	if SpaceMemberErr != nil {
		log.Fatalf("Error AutoMigrate SpaceMember %v", SpaceMemberErr)
	}

	EntitiesErr := db.AutoMigrate(&models.Entities{})
	if EntitiesErr != nil {
		log.Fatalf("Error AutoMigrate Entities %v", EntitiesErr)
	}

	PolGrpSubsEntityMappingErr := db.AutoMigrate(&models.PolGrpSubsEntityMapping{})
	if PolGrpSubsEntityMappingErr != nil {
		log.Fatalf("Error AutoMigrate PolGrpSubsEntityMapping %v", PolGrpSubsEntityMappingErr)
	}

	ShieldAppDomainMappingErr := db.AutoMigrate(&models.ShieldAppDomainMapping{})
	if ShieldAppDomainMappingErr != nil {
		log.Fatalf("Error AutoMigrate ShieldAppDomainMappingErr %v", ShieldAppDomainMappingErr)
	}

	createPgCrypto := db.Exec(`
	CREATE EXTENSION IF NOT EXISTS pgcrypto;`)

	if createPgCrypto.Error != nil {
		log.Fatal("Error")
	}

	resRandomBytes := db.Exec(`
	CREATE OR REPLACE FUNCTION public.gen_random_bytes(
		integer)
		RETURNS bytea
		LANGUAGE 'c'
		COST 1
		VOLATILE STRICT PARALLEL SAFE 
	AS '$libdir/pgcrypto', 'pg_random_bytes';
	
	ALTER FUNCTION public.gen_random_bytes(integer)
		OWNER TO postgres;
	`)

	if resRandomBytes.Error != nil {
		log.Fatal("Error")
	}

	resNanoId := db.Exec(`
	CREATE OR REPLACE FUNCTION public.nanoid(
		size integer DEFAULT 21)
		RETURNS text
		LANGUAGE 'plpgsql'
		COST 100
		STABLE PARALLEL UNSAFE
	AS $BODY$
	DECLARE
		id text := '';
		i int := 0;
		urlAlphabet char(64) := '_-0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz';
		bytes bytea := gen_random_bytes(size);
		byte int;
		pos int;
	BEGIN
		WHILE i < size LOOP
			byte := get_byte(bytes, i);
			pos := (byte & 63) + 1;
			id := id || substr(urlAlphabet, pos, 1);
			i = i + 1;
		END LOOP;
		RETURN id;
	END
	$BODY$;
	
	ALTER FUNCTION public.nanoid(integer)
		OWNER TO postgres;
	`)

	if resNanoId.Error != nil {
		log.Fatal("Error")
	}

	//seeding default app for which app needs to be managed using shield (use the same client_id in the login request)
	defaultPermissionsErr := db.Exec(`
		INSERT INTO permissions("permission_id","permission_name",description,category,mandatory,"created_at","updated_at") VALUES ('435e7c65-1fd7-4718-9944-69e90b520542','Name',NULL,NULL,'False','2022-02-11 13:57:06.069568+00','2022-02-11 13:57:06.069568+00') on conflict do nothing;
		INSERT INTO permissions("permission_id","permission_name",description,category,mandatory,"created_at","updated_at") VALUES ('5e417e6d-b24c-43e3-841f-d4f5ead70ff1','Address',NULL,NULL,'False','2022-02-11 13:57:06.838323+00','2022-02-11 13:57:06.838323+00') on conflict do nothing;
		INSERT INTO permissions("permission_id","permission_name",description,category,mandatory,"created_at","updated_at") VALUES ('4c6e1190-6049-4435-9292-5f58037b02fb','Phone',NULL,NULL,'False','2022-02-11 13:57:07.612243+00','2022-02-11 13:57:07.612243+00') on conflict do nothing;
		INSERT INTO permissions("permission_id","permission_name",description,category,mandatory,"created_at","updated_at") VALUES ('5194c89e-2f68-499f-a521-cb086f706ba8','Calendar','Access event details from your calendar app.',NULL,'False','2022-02-11 13:57:08.37531+00','2022-02-11 13:57:08.37531+00') on conflict do nothing;
		INSERT INTO permissions("permission_id","permission_name",description,category,mandatory,"created_at","updated_at") VALUES ('2feea3e9-2967-470a-892b-a578dea752d2','Location','Access your location while you are using the app.',NULL,'False','2022-02-11 13:57:09.139656+00','2022-02-11 13:57:09.139656+00') on conflict do nothing;
		INSERT INTO permissions("permission_id","permission_name",description,category,mandatory,"created_at","updated_at") VALUES ('0e9e649f-94e1-4c5c-b651-3a8729d97c15','WiFi Connection Info','Full network access.',NULL,'False','2022-02-11 13:57:09.91085+00','2022-02-11 13:57:09.91085+00') on conflict do nothing;
		INSERT INTO permissions("permission_id","permission_name",description,category,mandatory,"created_at","updated_at") VALUES ('fb69c178-e739-4c8a-a10b-a3ad490c23ae','Device ID',NULL,NULL,'False','2022-02-11 13:57:10.677277+00','2022-02-11 13:57:10.677277+00') on conflict do nothing;
		INSERT INTO permissions("permission_id","permission_name",description,category,mandatory,"created_at","updated_at") VALUES ('dfbd8da1-0eb5-4082-8bf9-bb676450758d','Can create and delete your appblox activity.',NULL,NULL,'False','2022-02-11 13:57:11.44848+00','2022-02-11 13:57:11.44848+00') on conflict do nothing;
		INSERT INTO permissions("permission_id","permission_name",description,category,mandatory,"created_at","updated_at") VALUES ('346a4734-42bb-4882-953a-3b3356d15248','Contacts',NULL,NULL,'False','2022-02-11 13:57:12.219896+00','2022-02-11 13:57:12.219896+00') on conflict do nothing;
		INSERT INTO permissions("permission_id","permission_name",description,category,mandatory,"created_at","updated_at") VALUES ('34f336bb-0b71-40bf-bbc8-fa9d9965a16c','Files',NULL,NULL,'False','2022-02-11 13:57:12.991628+00','2022-02-11 13:57:12.991628+00') on conflict do nothing;
		INSERT INTO permissions("permission_id","permission_name",description,category,mandatory,"created_at","updated_at") VALUES ('18d7bbbc-cc0e-4a99-a8d3-616eeff37344','Microphone',NULL,NULL,'False','2022-02-11 13:57:13.753364+00','2022-02-11 13:57:13.753364+00') on conflict do nothing;
		INSERT INTO permissions("permission_id","permission_name",description,category,mandatory,"created_at","updated_at") VALUES ('960e93c5-2e2f-4a23-80ab-dff7127bce2d','Email','Read. Compose. Send your emails.',NULL,'True','2022-02-11 13:57:04.262583+00','2022-02-11 13:57:04.262583+00') on conflict do nothing;
		INSERT INTO permissions("permission_id","permission_name",description,category,mandatory,"created_at","updated_at") VALUES ('ba74d677-7dfa-4e4c-80b4-b43619ab8bc5','Username',NULL,NULL,'True','2022-02-11 13:57:05.294117+00','2022-02-11 13:57:05.294117+00') on conflict do nothing ;
		`)

	if defaultPermissionsErr.Error != nil {
		log.Fatal("Error")
	}

	//seeding default app for which app needs to be managed using shield (use the same client_id in the login request)
	newApp := db.Exec(`
		INSERT INTO public.shield_apps(
			app_id, client_id, client_secret, app_name, app_sname, description, app_url, redirect_url, app_type, created_at, updated_at, deleted_at, owner_space_id, id)
			
			VALUES (nanoid(),'test-app-1','243db79e075ebec5b698783b77784442787cb10a60df60bcdb8c99d58bdaca794fff01d7a337b5093f7492b5b028832bc8115a67f7e4b379115b5e059c798485','test-app',
					'test-app','Test App ', 'http://localhost:3011','{http://localhost:3011}',2,now(),null,null,null,null) on conflict do nothing ;
		`)

	if newApp.Error != nil {
		log.Fatal("Error")
	}

	newAppPermissionErr := db.Exec(`
		INSERT INTO public.app_permissions(
			app_id, permission_id, mandatory, created_at, updated_at)
		select a.app_id,p.permission_id,p.mandatory,now(),null from shield_apps a inner join permissions p on true 
		where a.client_id='test-app-1' on conflict do nothing;
		`)

	if newAppPermissionErr.Error != nil {
		log.Fatal("Error")
	}

	// log.Fatal("DB migration completed!")
}
