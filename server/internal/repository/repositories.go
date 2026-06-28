package repository

import "gorm.io/gorm"

type Repositories struct {
	Admins      AdminRepository
	Users       UserRepository
	APIKeys     APIKeyRepository
	Wallets     WalletRepository
	Providers   ProviderRepository
	ImageModels ImageModelRepository
	ImageTasks  ImageTaskRepository
	Videos      VideoRepository
	Database    DatabaseRepository
	Tx          TransactionManager
}

func NewRepositories(db *gorm.DB) Repositories {
	return Repositories{
		Admins:      NewGormAdminRepository(db),
		Users:       NewGormUserRepository(db),
		APIKeys:     NewGormAPIKeyRepository(db),
		Wallets:     NewGormWalletRepository(db),
		Providers:   NewGormProviderRepository(db),
		ImageModels: NewGormImageModelRepository(db),
		ImageTasks:  NewGormImageTaskRepository(db),
		Videos:      NewGormVideoRepository(db),
		Database:    NewGormDatabaseRepository(db),
		Tx:          NewGormTransactionManager(db),
	}
}
