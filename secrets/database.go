package secrets

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/melardev/discord-message-protect/core"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

// for testing I use to enable this, so I use sqlite rather than
// MySQL which is what is being used in production
var useSqlite = false
var errorsDbConnect = 0

// CacheEntry - Is the unit in the cache, will contain the secret itself and some
// extra metadata
type CacheEntry struct {
	Secret *Secret
	// CachedAt Time when the secret has been cached
	CachedAt time.Time
	// LastRequestedAt - Last time the secret has been requested
	LastRequestedAt time.Time
}

type DbSecretManager struct {
	SecretMutex sync.Mutex
	// Cache contains entries that have been accessed recently
	// so they are ready to be served to next clients without making
	// a database lookup which would degrade the performance.
	Cache          map[string]*CacheEntry
	DeletedEntries int
}

func GetDatabase() (*gorm.DB, error) {

	if databaseCachedConnection != nil {
		db, err := databaseCachedConnection.DB()
		if err != nil {
			return nil, err
		}

		err = db.Ping()
		if err != nil {
			return nil, err
		}

		return databaseCachedConnection, nil
	}

	dbLock.Lock()

	if databaseCachedConnection != nil {

		db, err := databaseCachedConnection.DB()
		if err != nil {
			dbLock.Unlock()
			return nil, err
		}

		err = db.Ping()
		if err != nil {
			dbLock.Unlock()
			return nil, err
		}

		dbLock.Unlock()
		return databaseCachedConnection, nil
	}

	var err error

	// databaseCachedConnection, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if useSqlite {
		fmt.Printf("Using SQlite Database\n")
		databaseCachedConnection, err = gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
		dbLock.Unlock()
		return databaseCachedConnection, err
	} else {
		log.Printf("Connecting to MySQL\n")
		host := "db"
		user := "root"
		password := "example"
		dbName := "discord_protect"
		dbPort := 3306

		if len(os.Getenv("DB_HOST")) > 0 {
			host = os.Getenv("DB_HOST")
		}

		if len(os.Getenv("DB_USER")) > 0 {
			user = os.Getenv("DB_USER")
		}

		if len(os.Getenv("DB_PORT")) > 0 {
			dbPort, err = strconv.Atoi(os.Getenv("DB_PORT"))
			if err != nil {
				dbLock.Unlock()
				return nil, err
			}
		}

		if len(os.Getenv("DB_PASSWORD")) > 0 {
			password = os.Getenv("DB_PASSWORD")
		}

		dbUri := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			user, password, host, dbPort, dbName)

		databaseCachedConnection, err = gorm.Open(mysql.Open(dbUri), &gorm.Config{})
		if err != nil {
			errorsDbConnect++
			log.Printf("[-] Database - An error occurred Connecting to MySQL %d - %v\n", errorsDbConnect, err)
			databaseCachedConnection = nil
			if errorsDbConnect >= 3 {
				useSqlite = true
				log.Printf("Database - Falling back to SQLite")
			} else {
				log.Printf("Database - retrying again in 3 seconds")
				// Wait a little, the database may have not been started yet.
				time.Sleep(time.Second * 3)
			}
			dbLock.Unlock()
			return GetDatabase()
		} else {
			log.Printf("[+] Database - Connection successfully made to MySQL\n")
		}
		dbLock.Unlock()

		return databaseCachedConnection, err
	}
}

var databaseCachedConnection *gorm.DB

func Migrate(database *gorm.DB) {

	var err error

	err = database.AutoMigrate(&Secret{})
	if err != nil {
		panic(err)
	}

}

func (s *DbSecretManager) Delete(secretId string) error {
	s.SecretMutex.Lock()
	if _, found := s.Cache[secretId]; found {
		delete(s.Cache, secretId)
		s.DeletedEntries++
	}

	database, err := GetDatabase()
	if err != nil {
		return err
	}

	result := database.Where("secret_id = ?", secretId).Delete(&Secret{})

	if err = result.Error; err != nil {
		return err
	}

	if result.RowsAffected == 0 {
		// Log weir behaviour
	}

	s.SecretMutex.Unlock()
	return nil
}

func NewDbSecretManager(*core.Config) *DbSecretManager {
	d := &DbSecretManager{
		Cache: map[string]*CacheEntry{},
	}

	db, err := GetDatabase()
	if err != nil {
		panic(err)
	}

	Migrate(db)
	go d.DoMaintenance()
	return d
}

var dbLock = &sync.Mutex{}

func (s *DbSecretManager) GetById(id string) (*Secret, error) {

	s.SecretMutex.Lock()
	if entry, found := s.Cache[id]; found {
		entry.LastRequestedAt = time.Now().UTC()
		s.SecretMutex.Unlock()
		return entry.Secret, nil
	} else {

		database, err := GetDatabase()
		if err != nil {
			s.SecretMutex.Unlock()
			return nil, err
		}

		var secret *Secret
		if err = database.Raw("select * from secrets where secret_id = ?", id).First(&secret).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				s.SecretMutex.Unlock()
				return nil, nil
			}
			s.SecretMutex.Unlock()
			return nil, err
		} else {
			now := time.Now().UTC()
			s.Cache[id] = &CacheEntry{
				Secret:          secret,
				CachedAt:        now,
				LastRequestedAt: now,
			}
			s.SecretMutex.Unlock()
			return secret, nil
		}
	}
}

func (s *DbSecretManager) Create(dto *CreateSecretDto) (*Secret, error) {
	now := time.Now().UTC()

	var secret *Secret
	s.SecretMutex.Lock()
	secret = &Secret{
		SecretId: uuid.Must(uuid.NewRandom()).String(),
		User: &core.DiscordUser{
			Id:       dto.User.Discriminator,
			Username: dto.User.Username,
		},
		UserName:  fmt.Sprintf("%s#%s", dto.User.Username, dto.User.Discriminator),
		Message:   dto.Message,
		ChannelId: dto.ChannelId,
		ImageUrl:  dto.ImageUrl,
		CreatedAt: now,
		UpdatedAt: now,
	}

	database, err := GetDatabase()
	if err != nil {
		s.SecretMutex.Unlock()
		return nil, err
	}

	if err = database.Create(&secret).Error; err != nil {
		s.SecretMutex.Unlock()
		return nil, err
	}

	s.Cache[secret.SecretId] = &CacheEntry{
		Secret:          secret,
		CachedAt:        now,
		LastRequestedAt: now,
	}

	s.SecretMutex.Unlock()

	return secret, nil
}

func (s *DbSecretManager) Update(secret *Secret) error {
	database, err := GetDatabase()
	if err != nil {
		return err
	}

	if err = database.Save(secret).Error; err != nil {
		return err
	}
	return nil
}

func (s *DbSecretManager) UpdateMessageId(secret *Secret) error {
	database, err := GetDatabase()
	if err != nil {
		return err
	}

	if err = database.Model(secret).Update("message_id", secret.MessageId).Error; err != nil {
		return err
	}
	return nil
}

// DoMaintenance - Will clean cache entries if they were not used for some time.
func (s *DbSecretManager) DoMaintenance() {

	time.Sleep(time.Second * 3)
	for {
		if s.SecretMutex.TryLock() {
			now := time.Now().UTC()
			for key, entry := range s.Cache {
				if now.Sub(entry.LastRequestedAt) > time.Minute*15 {
					delete(s.Cache, key)
					s.DeletedEntries++
				}
			}

			// Rotate the map, golang does not rotate the maps, so as we grow it,
			// even if we delete some entries, the map is never gonna shrink in size
			// to avoid having memory problems, rotate it manually
			if s.DeletedEntries >= 100 {
				newMap := make(map[string]*CacheEntry, len(s.Cache))
				for k, v := range s.Cache {
					newMap[k] = v
				}
				s.Cache = newMap
			}
			s.SecretMutex.Unlock()
		}

		time.Sleep(time.Second * 10)
	}
}
