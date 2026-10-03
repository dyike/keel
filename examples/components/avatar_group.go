package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("avatar_group", "controls", func() core.Widget {
		avatars := []*kit.AvatarView{kit.Avatar("张三").Status(kit.AvatarOnline), kit.Avatar("Ada Lovelace"), kit.Avatar("Bob"), kit.Avatar("Carol"), kit.Avatar("Dave")}
		small := kit.AvatarGroup(avatars...).Size(24)
		limited := kit.AvatarGroup(avatars...).Limit(3)
		ellipsis := kit.AvatarGroup(avatars...).Size(56).Limit(2).Ellipsis(true)
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(theme.SpaceXl).Gap(theme.SpaceLg).Items(el.Start).Child(
				el.Text("团队成员"), small.Render(cx), el.Text("显示 3 人，其余用数量表示"), limited.Render(cx), el.Text("省略号模式"), ellipsis.Render(cx))
		}))
	})
}
