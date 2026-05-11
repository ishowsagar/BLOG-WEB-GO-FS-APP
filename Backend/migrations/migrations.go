package migrations

import (
	"log/slog"

	dumper "github.com/goforj/godump"
	"github.com/ishowsagar/go-blog-web-application/models"
	"gorm.io/gorm"
)

// tables creations and auto migrations of the tables
func AutoMigrate(db *gorm.DB) error {

	err := db.AutoMigrate(&models.Comment{},&models.Post{},&models.User{},&models.Token{},&models.Like{},&models.Follow{},) // takes in concrete parent type not fields injected
	if err != nil {
		return err
	}

	// err = EnsureCascadeConstraints(db)
	// if err != nil {
	// 	return err
	// }

	dumper.Dump(models.Comment{},models.Post{},models.User{},models.Token{},models.Like{},models.Follow{})
	slog.Info("successfully migrated models🚀.")
	return nil
}

func Demigrate(db *gorm.DB) error {
	err := db.Migrator().DropColumn(&models.Like{},"time_stamp")

	if err != nil {
		return err
	}

	slog.Info("successfully demigrated.")
	return nil
}

// fixed - constraint that were not setup but tables created without them
func EnsureCascadeConstraints(db *gorm.DB) error {
    // Comments -> Post, User
    if db.Migrator().HasConstraint(&models.Comment{}, "Post") {
        if err := db.Migrator().DropConstraint(&models.Comment{}, "Post"); err != nil {
            return err
        }
    }
    if err := db.Migrator().CreateConstraint(&models.Comment{}, "Post"); err != nil {
        return err
    }

    if db.Migrator().HasConstraint(&models.Comment{}, "User") {
        _ = db.Migrator().DropConstraint(&models.Comment{}, "User")
    }
    if err := db.Migrator().CreateConstraint(&models.Comment{}, "User"); err != nil {
        return err
    }

    // Likes -> Post, User
    if db.Migrator().HasConstraint(&models.Like{}, "Post") {
        _ = db.Migrator().DropConstraint(&models.Like{}, "Post")
    }
    if err := db.Migrator().CreateConstraint(&models.Like{}, "Post"); err != nil {
        return err
    }
    if db.Migrator().HasConstraint(&models.Like{}, "User") {
        _ = db.Migrator().DropConstraint(&models.Like{}, "User")
    }
    if err := db.Migrator().CreateConstraint(&models.Like{}, "User"); err != nil {
        return err
    }

    // Tokens -> User
    if db.Migrator().HasConstraint(&models.Token{}, "User") {
        _ = db.Migrator().DropConstraint(&models.Token{}, "User")
    }
    if err := db.Migrator().CreateConstraint(&models.Token{}, "User"); err != nil {
        return err
    }

	slog.Info("added constraints concisely")
    return nil
}