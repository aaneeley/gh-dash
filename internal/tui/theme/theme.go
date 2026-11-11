package theme

import (
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"
	"charm.land/log/v2"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
)

type Theme struct {
	// Accent is the color used to highlight what is focused or selected.
	Accent compat.AdaptiveColor
	// Secondary and Tertiary are used for states, e.g. open and merged PRs.
	Secondary compat.AdaptiveColor
	Tertiary  compat.AdaptiveColor
	// Background is the color of the whole screen.
	Background compat.AdaptiveColor
	// PanelBackground is the color of raised areas such as the preview pane.
	PanelBackground compat.AdaptiveColor
	// ElementBackground is the color of blocks such as inputs and comments.
	ElementBackground compat.AdaptiveColor

	SelectedBackground      compat.AdaptiveColor // config.Theme.Colors.Background.Selected
	PrimaryBorder           compat.AdaptiveColor // config.Theme.Colors.Border.Primary
	FaintBorder             compat.AdaptiveColor // config.Theme.Colors.Border.Faint
	SecondaryBorder         compat.AdaptiveColor // config.Theme.Colors.Border.Secondary
	FaintText               compat.AdaptiveColor // config.Theme.Colors.Text.Faint
	PrimaryText             compat.AdaptiveColor // config.Theme.Colors.Text.Primary
	SecondaryText           compat.AdaptiveColor // config.Theme.Colors.Text.Secondary
	InvertedText            compat.AdaptiveColor // config.Theme.Colors.Text.Inverted
	SuccessText             compat.AdaptiveColor // config.Theme.Colors.Text.Success
	WarningText             compat.AdaptiveColor // config.Theme.Colors.Text.Warning
	ErrorText               compat.AdaptiveColor // config.Theme.Colors.Text.Error
	ActorText               compat.AdaptiveColor // config.Theme.Colors.Text.Actor
	NewContributorIconColor compat.AdaptiveColor // config.Theme.Colors.Icon.NewContributor
	ContributorIconColor    compat.AdaptiveColor // config.Theme.Colors.Icon.Contributor
	CollaboratorIconColor   compat.AdaptiveColor // config.Theme.Colors.Icon.Collaborator
	MemberIconColor         compat.AdaptiveColor // config.Theme.Colors.Icon.Member
	OwnerIconColor          compat.AdaptiveColor // config.Theme.Colors.Icon.Owner
	UnknownRoleIconColor    compat.AdaptiveColor // config.Theme.Colors.Icon.UnknownRole
	NewContributorIcon      string               // config.Theme.Icons.NewContributor
	ContributorIcon         string               // config.Theme.Icons.Contributor
	CollaboratorIcon        string               // config.Theme.Icons.Collaborator
	MemberIcon              string               // config.Theme.Icons.Member
	OwnerIcon               string               // config.Theme.Icons.Owner
	UnknownRoleIcon         string               // config.Theme.Icons.UnknownRole
}

// The default colors are those of opencode's Catppuccin theme: Mocha on dark
// terminals and Latte on light ones. The comments name the color of the
// Catppuccin palette and, where there is one, the role it has in opencode.
func adaptive(light, dark string) compat.AdaptiveColor {
	return compat.AdaptiveColor{Light: lipgloss.Color(light), Dark: lipgloss.Color(dark)}
}

var DefaultTheme = &Theme{
	Accent:             adaptive("#1e66f5", "#89b4fa"), // blue (primary)
	Secondary:          adaptive("#8839ef", "#cba6f7"), // mauve (secondary)
	Tertiary:           adaptive("#ea76cb", "#f5c2e7"), // pink (accent)
	Background:         adaptive("#eff1f5", "#1e1e2e"), // base (background)
	PanelBackground:    adaptive("#e6e9ef", "#181825"), // mantle (backgroundPanel)
	ElementBackground:  adaptive("#dce0e8", "#11111b"), // crust (backgroundElement)
	SelectedBackground: adaptive("#ccd0da", "#313244"), // surface0
	PrimaryBorder:      adaptive("#bcc0cc", "#45475a"), // surface1 (borderActive)
	SecondaryBorder:    adaptive("#acb0be", "#585b70"), // surface2 (borderSubtle)
	FaintBorder:        adaptive("#ccd0da", "#313244"), // surface0 (border)
	PrimaryText:        adaptive("#4c4f69", "#cdd6f4"), // text
	SecondaryText:      adaptive("#5c5f77", "#bac2de"), // subtext1
	FaintText:          adaptive("#7c7f93", "#9399b2"), // overlay2 (textMuted)
	InvertedText:       adaptive("#eff1f5", "#11111b"), // base / crust
	SuccessText:        adaptive("#40a02b", "#a6e3a1"), // green (success)
	WarningText:        adaptive("#df8e1d", "#f9e2af"), // yellow (warning)
	ErrorText:          adaptive("#d20f39", "#f38ba8"), // red (error)
	ActorText:          adaptive("#5c5f77", "#bac2de"), // Same as SecondaryText

	NewContributorIconColor: adaptive("#40a02b", "#a6e3a1"), // green
	ContributorIconColor:    adaptive("#209fb5", "#74c7ec"), // sapphire
	CollaboratorIconColor:   adaptive("#fe640b", "#fab387"), // peach
	MemberIconColor:         adaptive("#fe640b", "#fab387"),
	OwnerIconColor:          adaptive("#fe640b", "#fab387"),
	UnknownRoleIconColor:    adaptive("#fe640b", "#fab387"),

	NewContributorIcon: constants.NewContributorIcon,
	ContributorIcon:    constants.ContributorIcon,
	CollaboratorIcon:   constants.CollaboratorIcon,
	MemberIcon:         constants.MemberIcon,
	OwnerIcon:          constants.OwnerIcon,
	UnknownRoleIcon:    constants.UnknownRoleIcon,
}

func ParseTheme(cfg *config.Config) Theme {
	_shimColor := func(color config.Color, fallback compat.AdaptiveColor) compat.AdaptiveColor {
		if color == "" {
			return fallback
		}
		return compat.AdaptiveColor{
			Light: lipgloss.Color(string(color)),
			Dark:  lipgloss.Color(string(color)),
		}
	}
	_shimIcon := func(icon string, fallback string) string {
		if icon != "" {
			return icon
		}
		return fallback
	}

	if cfg.Theme.Colors != nil {
		DefaultTheme.SelectedBackground = _shimColor(
			cfg.Theme.Colors.Inline.Background.Selected,
			DefaultTheme.SelectedBackground,
		)
		DefaultTheme.PrimaryBorder = _shimColor(
			cfg.Theme.Colors.Inline.Border.Primary,
			DefaultTheme.PrimaryBorder,
		)
		DefaultTheme.FaintBorder = _shimColor(
			cfg.Theme.Colors.Inline.Border.Faint,
			DefaultTheme.FaintBorder,
		)
		DefaultTheme.SecondaryBorder = _shimColor(
			cfg.Theme.Colors.Inline.Border.Secondary,
			DefaultTheme.SecondaryBorder,
		)
		DefaultTheme.FaintText = _shimColor(
			cfg.Theme.Colors.Inline.Text.Faint,
			DefaultTheme.FaintText,
		)
		DefaultTheme.PrimaryText = _shimColor(
			cfg.Theme.Colors.Inline.Text.Primary,
			DefaultTheme.PrimaryText,
		)
		DefaultTheme.SecondaryText = _shimColor(
			cfg.Theme.Colors.Inline.Text.Secondary,
			DefaultTheme.SecondaryText,
		)
		DefaultTheme.InvertedText = _shimColor(
			cfg.Theme.Colors.Inline.Text.Inverted,
			DefaultTheme.InvertedText,
		)
		DefaultTheme.SuccessText = _shimColor(
			cfg.Theme.Colors.Inline.Text.Success,
			DefaultTheme.SuccessText,
		)
		DefaultTheme.WarningText = _shimColor(
			cfg.Theme.Colors.Inline.Text.Warning,
			DefaultTheme.WarningText,
		)
		log.Debug("error text",
			"cfg",
			cfg.Theme.Colors.Inline.Text.Error,
			"default",
			DefaultTheme.ErrorText,
		)
		DefaultTheme.ErrorText = _shimColor(
			cfg.Theme.Colors.Inline.Text.Error,
			DefaultTheme.ErrorText,
		)
		DefaultTheme.ActorText = _shimColor(
			cfg.Theme.Colors.Inline.Text.Actor,
			DefaultTheme.ActorText,
		)
		DefaultTheme.NewContributorIconColor = _shimColor(
			cfg.Theme.Colors.Inline.Icon.NewContributor,
			DefaultTheme.NewContributorIconColor,
		)
		DefaultTheme.ContributorIconColor = _shimColor(
			cfg.Theme.Colors.Inline.Icon.Contributor,
			DefaultTheme.ContributorIconColor,
		)
		DefaultTheme.CollaboratorIconColor = _shimColor(
			cfg.Theme.Colors.Inline.Icon.Collaborator,
			DefaultTheme.CollaboratorIconColor,
		)
		DefaultTheme.MemberIconColor = _shimColor(
			cfg.Theme.Colors.Inline.Icon.Member,
			DefaultTheme.MemberIconColor,
		)
		DefaultTheme.OwnerIconColor = _shimColor(
			cfg.Theme.Colors.Inline.Icon.Owner,
			DefaultTheme.OwnerIconColor,
		)
	}

	if cfg.ShowAuthorIcons && cfg.Theme.Icons != nil {
		DefaultTheme.NewContributorIcon = _shimIcon(
			cfg.Theme.Icons.Inline.NewContributor,
			DefaultTheme.NewContributorIcon,
		)
		DefaultTheme.ContributorIcon = _shimIcon(
			cfg.Theme.Icons.Inline.Contributor,
			DefaultTheme.ContributorIcon,
		)
		DefaultTheme.CollaboratorIcon = _shimIcon(
			cfg.Theme.Icons.Inline.Collaborator,
			DefaultTheme.CollaboratorIcon,
		)
		DefaultTheme.MemberIcon = _shimIcon(
			cfg.Theme.Icons.Inline.Member,
			DefaultTheme.MemberIcon,
		)
		DefaultTheme.OwnerIcon = _shimIcon(
			cfg.Theme.Icons.Inline.Owner,
			DefaultTheme.OwnerIcon,
		)
		DefaultTheme.UnknownRoleIcon = _shimIcon(
			cfg.Theme.Icons.Inline.UnknownRole,
			DefaultTheme.UnknownRoleIcon,
		)
	}

	return *DefaultTheme
}
