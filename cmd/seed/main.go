package main

import (
	"log"

	"github.com/kidkender/prohealth-api/internal/config"
	"github.com/kidkender/prohealth-api/internal/model"
	"github.com/kidkender/prohealth-api/pkg/database"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var departments = []model.Department{
	{Name: "Emergency Department"},
	{Name: "Obstetrics & Gynecology"},
	{Name: "Pediatric Department"},
	{Name: "Cardiology Department"},
	{Name: "Pulmonology Department"},
	{Name: "Psychiatry Department"},
	{Name: "Neurology Department"},
}

type doctorSeed struct {
	email      string
	fullname   string
	phone      string
	department string
}

var doctors = []doctorSeed{
	{"lisachen@prohealth.com", "Dr. Lisa Chen, MD", "+123-456-7890", "Emergency Department"},
	{"michaeljohnson@prohealth.com", "Dr. Michael Johnson, MD", "+123-456-7891", "Emergency Department"},
	{"karenlee@prohealth.com", "Dr. Karen Lee, MD", "+123-456-7892", "Obstetrics & Gynecology"},
	{"sarahpatel@prohealth.com", "Dr. Sarah Patel, MD", "+123-456-7893", "Pediatric Department"},
	{"davidnguyen@prohealth.com", "Dr. David Nguyen, MD", "+123-456-7894", "Pediatric Department"},
	{"susanbones@prohealth.com", "Dr. Susan Bones, MD", "+123-456-7895", "Pediatric Department"},
	{"jameslee@prohealth.com", "Dr. James Lee, MD", "+123-456-7896", "Cardiology Department"},
	{"michellekim@prohealth.com", "Dr. Michelle Kim, MD", "+123-456-7897", "Cardiology Department"},
	{"johnsmith@prohealth.com", "Dr. John Smith, MD", "+123-456-7898", "Pulmonology Department"},
	{"sarahlee@prohealth.com", "Dr. Sarah Lee, MD, MPH, FAPA", "+123-456-7899", "Psychiatry Department"},
	{"davidpark@prohealth.com", "Dr. David Park, MD", "+123-456-7900", "Psychiatry Department"},
	{"mariarodriguez@prohealth.com", "Dr. Maria Rodriguez, MD", "+123-456-7901", "Neurology Department"},
}

func main() {
	config.LoadEnv()

	db, err := database.NewPostgresDB()
	if err != nil {
		log.Fatal("failed to connect database: ", err)
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&departments).Error; err != nil {
			return err
		}
		log.Printf("Seeded %d departments", len(departments))

		var deptRows []model.Department
		if err := tx.Find(&deptRows).Error; err != nil {
			return err
		}
		deptMap := make(map[string]uint, len(deptRows))
		for _, d := range deptRows {
			deptMap[d.Name] = d.ID
		}

		hospitalSeeds := []model.Hospital{
			{Name: "ProHealth General Hospital", Slug: "prohealth-general", Description: "Main general hospital", Address: "123 Medical Street"},
			{Name: "ProHealth City Clinic", Slug: "prohealth-city", Description: "City branch clinic", Address: "456 City Avenue"},
		}
		for i := range hospitalSeeds {
			if err := tx.Where(model.Hospital{Slug: hospitalSeeds[i].Slug}).FirstOrCreate(&hospitalSeeds[i]).Error; err != nil {
				return err
			}
		}
		hospital := hospitalSeeds[0]
		log.Printf("Seeded %d hospitals", len(hospitalSeeds))

		adminHash, err := bcrypt.GenerateFromPassword([]byte("Admin@123"), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		admin := model.User{Email: "admin@prohealth.com", Password: string(adminHash), SystemRole: model.SystemRoleAdmin}
		if res := tx.Where(model.User{Email: admin.Email}).FirstOrCreate(&admin); res.Error != nil {
			return res.Error
		}
		tx.Where(model.UserDetail{UserId: admin.ID}).FirstOrCreate(&model.UserDetail{
			UserId: admin.ID, Fullname: "Admin",
		})
		log.Println("Admin seeded")

		hospitalAdminHash, err := bcrypt.GenerateFromPassword([]byte("HospAdmin@123"), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		hospitalAdmin := model.User{Email: "hospitaladmin@prohealth.com", Password: string(hospitalAdminHash), SystemRole: model.SystemRoleUser}
		if res := tx.Where(model.User{Email: hospitalAdmin.Email}).FirstOrCreate(&hospitalAdmin); res.Error != nil {
			return res.Error
		}
		tx.Where(model.UserDetail{UserId: hospitalAdmin.ID}).FirstOrCreate(&model.UserDetail{
			UserId: hospitalAdmin.ID, Fullname: "Hospital Admin",
		})
		tx.Where(model.HospitalMember{HospitalID: hospital.ID, UserID: hospitalAdmin.ID}).
			FirstOrCreate(&model.HospitalMember{
				HospitalID: hospital.ID,
				UserID:     hospitalAdmin.ID,
				Role:       model.RoleAdmin,
			})
		log.Println("Hospital admin seeded")

		// Seed doctors
		doctorHash, err := bcrypt.GenerateFromPassword([]byte("Doctor@123"), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		multiHospitalDoctors := map[string]bool{
			"lisachen@prohealth.com": true,
			"jameslee@prohealth.com": true,
		}

		for _, d := range doctors {
			user := model.User{Email: d.email, Password: string(doctorHash), SystemRole: model.SystemRoleUser}
			if res := tx.Where(model.User{Email: d.email}).FirstOrCreate(&user); res.Error != nil {
				return res.Error
			}
			deptID := deptMap[d.department]
			tx.Where(model.UserDetail{UserId: user.ID}).FirstOrCreate(&model.UserDetail{
				UserId: user.ID, Fullname: d.fullname, PhoneNumber: d.phone, DepartmentID: &deptID,
			})

			joinHospitals := []model.Hospital{hospital}
			if multiHospitalDoctors[d.email] {
				joinHospitals = hospitalSeeds
			}
			for _, h := range joinHospitals {
				tx.Where(model.HospitalMember{HospitalID: h.ID, UserID: user.ID}).
					FirstOrCreate(&model.HospitalMember{
						HospitalID: h.ID,
						UserID:     user.ID,
						Role:       model.RoleDoctor,
					})
			}
		}
		log.Printf("Seeded %d doctors", len(doctors))

		return nil
	})

	if err != nil {
		log.Fatal("seed failed: ", err)
	}
	log.Println("All seeds completed successfully")
}
