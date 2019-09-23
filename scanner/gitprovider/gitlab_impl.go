package gitprovider

import (
	"errors"
	"github.com/xanzy/go-gitlab"
	"strconv"
)

type GitlabProvider struct {
	Client *gitlab.Client
	AdditionalParams map[string]string
	token string
}

func (g *GitlabProvider) Initialize(baseURL, token string, additionalParams map[string]string) error {
	if !g.ValidateAdditionalParams(additionalParams) {
		return ErrInvalidAdditionalParams
	}

	g.token = token
	g.AdditionalParams = additionalParams
	g.Client = gitlab.NewClient(nil, token)
	err := g.Client.SetBaseURL(baseURL)
	if err != nil {
		return err
	}

	return nil
}

func (g *GitlabProvider) GetRepository(opt map[string]string) (*Repository, error) {
	id, exists := opt["id"]
	if !exists {
		return nil, errors.New("id option does not exists in map")
	}
	proj, _, err := g.Client.Projects.GetProject(id, nil)
	if err != nil {
		return nil, err
	}

	repo := &Repository{
		ID:            strconv.Itoa( proj.ID),
		Name:          proj.Name,
		FullName:      proj.Name,
		CloneURL:      proj.SSHURLToRepo,
		URL:           proj.WebURL,
		DefaultBranch: proj.DefaultBranch,
		Description:   proj.Description,
		Homepage:      proj.WebURL,
		Owner:         "",
	}
	return repo, nil
}

func (g *GitlabProvider) ValidateAdditionalParams(additionalParams map[string]string) bool {
	return true
}

func (g *GitlabProvider) Name() string {
	return GitlabName
}
