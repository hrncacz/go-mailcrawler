package config

// Level 0
type Config struct {
	Email Email
	Ftp   Ftp
}

// Level 1
type Email struct {
	Auth               AuthEmail
	Paths              Paths
	DownloadFolder     string
	DaysToProcessFiles int
}

type Ftp struct {
	Auth           AuthFtp
	UploadFolder   string
	DownloadFolder string
	UploadArchive  string
	RemoteUp       string
	RemoteDown     string
}

// Level 2
type AuthEmail struct {
	User     string
	Password string
	Host     string
	Port     int
	Tls      bool
}
type AuthFtp struct {
	Host     string
	Port     string
	Secure   bool
	User     string
	Password string
}
type Paths struct {
	Source         string
	ProcessedMails string
	InvalidMails   string
}
