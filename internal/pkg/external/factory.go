package external

import (
	"sync"

	"github.com/rayda/rayda-service/internal/config"
	"github.com/rayda/rayda-service/internal/pkg/external/communication"
	"github.com/rayda/rayda-service/internal/pkg/external/payment"
	"github.com/rayda/rayda-service/internal/pkg/external/usermanagement"
)

// Services holds all external service clients
type Services struct {
	UserManagement *usermanagement.MockClient
	Payment       *payment.MockClient
	Communication *communication.MockClient
}

var (
	once     sync.Once
	instance *Services
)

// InitServices initializes all external service clients
func InitServices(cfg *config.Config) *Services {
	once.Do(func() {
		instance = &Services{
			UserManagement: usermanagement.NewMockClient(),
			Payment:       payment.NewMockClient(),
			Communication: communication.NewMockClient(),
		}
	})

	return instance
}

// GetServices returns the singleton instance of external services
func GetServices() *Services {
	if instance == nil {
		panic("external services not initialized - call InitServices first")
	}
	return instance
}

// Shutdown gracefully shuts down all external service clients
func (s *Services) Shutdown() error {
	// Add any cleanup logic for external services here
	return nil
}
