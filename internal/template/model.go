package template

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound              = errors.New("template not found")
	ErrInvalidID             = errors.New("invalid template id")
	ErrNameConflict          = errors.New("template name already exists")
	ErrUnresolvedPlaceholder = errors.New("unresolved template variables")
)

var placeholderRegex = regexp.MustCompile(`\{\{[^}]+\}\}`)

type Template struct {
	ID        uuid.UUID
	Name      string
	Body      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (t *Template) RenderBody(vars map[string]string) (string, error) {
	body := t.Body
	for k, v := range vars {
		body = strings.ReplaceAll(body, fmt.Sprintf("{{%s}}", k), v)
	}
	if placeholderRegex.MatchString(body) {
		unresolved := placeholderRegex.FindAllString(body, -1)
		return "", fmt.Errorf("%w: %s", ErrUnresolvedPlaceholder, strings.Join(unresolved, ", "))
	}
	return body, nil
}
