package reorder

import (
	"context"
	"strings"

	"github.com/aviator-co/av/internal/git"
	"github.com/aviator-co/av/internal/utils/colors"
	"github.com/aviator-co/av/internal/utils/errutils"
	"github.com/kr/text"
)

// PickCmd is a command that picks a commit from the history and applies it on
// top of the current HEAD.
type PickCmd struct {
	Commit  string
	Comment string
}

func (p PickCmd) Execute(ctx *Context) error {
	err := ctx.Repo.CherryPick(context.Background(), git.CherryPick{
		Commits: []string{p.Commit},
		// Use FastForward to avoid always amending commits.
		FastForward: true,
	})
	if conflict, ok := errutils.As[git.ErrCherryPickConflict](err); ok {
		ctx.Print(
			colors.Failure("  - ", conflict.Error(), "\n"),
			colors.Faint(text.Indent(strings.TrimRight(conflict.Output, "\n"), "        "), "\n"),
		)
		return ErrInterruptReorder
	} else if err != nil {
		return err
	}

	head, err := ctx.Repo.RevParse(context.Background(), &git.RevParse{Rev: "HEAD"})
	if err != nil {
		return err
	}
	ctx.Print(
		colors.Success("  - applied commit "),
		colors.UserInput(git.ShortSha(p.Commit)),
		colors.Success(" without conflict (HEAD is now at "),
		colors.UserInput(git.ShortSha(head)),
		colors.Success(")\n"),
	)
	ctx.State.Head = head
	return nil
}

func (p PickCmd) String() string {
	sb := strings.Builder{}
	sb.WriteString("pick ")
	sb.WriteString(p.Commit)
	if p.Comment != "" {
		sb.WriteString("  # ")
		sb.WriteString(p.Comment)
	}
	return sb.String()
}

var _ Cmd = &PickCmd{}

func parsePickCmd(args []string) (Cmd, error) {
	if len(args) != 1 {
		return nil, ErrInvalidCmd{"pick", "exactly one argument is required (the commit to pick)"}
	}
	return PickCmd{
		Commit: args[0],
	}, nil
}
