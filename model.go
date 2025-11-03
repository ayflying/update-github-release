package update

import "time"

// 对应 GitHub API 响应的核心字段
type GitHubRelease struct {
	Url             string     `json:"url"`
	AssetsUrl       string     `json:"assets_url"`
	UploadUrl       string     `json:"upload_url"`
	HtmlUrl         string     `json:"html_url"`
	Id              int        `json:"id"`
	Author          *Author    `json:"author"`
	TagName         string     `json:"tag_name"`
	Assets          []*Assets  `json:"assets"`
	NodeId          string     `json:"node_id"`
	TargetCommitish string     `json:"target_commitish"`
	Name            string     `json:"name"`
	Draft           bool       `json:"draft"`
	Immutable       bool       `json:"immutable"`
	Prerelease      bool       `json:"prerelease"`
	CreatedAt       *time.Time `json:"created_at"`
	UpdatedAt       *time.Time `json:"updated_at"`
	PublishedAt     *time.Time `json:"published_at"`
	TarballUrl      string     `json:"tarball_url"`
	ZipballUrl      string     `json:"zipball_url"`
	Body            string     `json:"body"`
}

type Author struct {
	Login             string `json:"login"`
	Id                int    `json:"id"`
	NodeId            string `json:"node_id"`
	AvatarUrl         string `json:"avatar_url"`
	GravatarId        string `json:"gravatar_id"`
	Url               string `json:"url"`
	HtmlUrl           string `json:"html_url"`
	FollowersUrl      string `json:"followers_url"`
	FollowingUrl      string `json:"following_url"`
	GistsUrl          string `json:"gists_url"`
	StarredUrl        string `json:"starred_url"`
	SubscriptionsUrl  string `json:"subscriptions_url"`
	OrganizationsUrl  string `json:"organizations_url"`
	ReposUrl          string `json:"repos_url"`
	EventsUrl         string `json:"events_url"`
	ReceivedEventsUrl string `json:"received_events_url"`
	Type              string `json:"type"`
	UserViewType      string `json:"user_view_type"`
	SiteAdmin         bool   `json:"site_admin"`
}

type Assets struct {
	Url                string    `json:"url"`
	Id                 int       `json:"id"`
	NodeId             string    `json:"node_id"`
	Name               string    `json:"name"`
	Label              string    `json:"label"`
	Uploader           *Uploader `json:"uploader"`
	ContentType        string    `json:"content_type"`
	State              string    `json:"state"`
	Size               int       `json:"size"`
	Digest             string    `json:"digest"`
	DownloadCount      int       `json:"download_count"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	BrowserDownloadUrl string    `json:"browser_download_url"`
}

type Uploader struct {
	Login             string `json:"login"`
	Id                int    `json:"id"`
	NodeId            string `json:"node_id"`
	AvatarUrl         string `json:"avatar_url"`
	GravatarId        string `json:"gravatar_id"`
	Url               string `json:"url"`
	HtmlUrl           string `json:"html_url"`
	FollowersUrl      string `json:"followers_url"`
	FollowingUrl      string `json:"following_url"`
	GistsUrl          string `json:"gists_url"`
	StarredUrl        string `json:"starred_url"`
	SubscriptionsUrl  string `json:"subscriptions_url"`
	OrganizationsUrl  string `json:"organizations_url"`
	ReposUrl          string `json:"repos_url"`
	EventsUrl         string `json:"events_url"`
	ReceivedEventsUrl string `json:"received_events_url"`
	Type              string `json:"type"`
	UserViewType      string `json:"user_view_type"`
	SiteAdmin         bool   `json:"site_admin"`
}
