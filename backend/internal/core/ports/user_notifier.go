package ports

type UserNotifier interface {
	NotifyRegistrationIntent(to, code string) error
	NotifyRegistrationSuccess(to string) error
}
