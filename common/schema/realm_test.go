package schema

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestRealm_FilterByVisibility(t *testing.T) {
	t.Run("filters out lower visibility", func(t *testing.T) {
		realm := &Realm{
			Name: "test",
			Modules: []*Module{{
				Name:  "test",
				Decls: []Decl{&Verb{Name: "test", Visibility: VisibilityScopeNone}},
			}},
		}
		filtered := realm.FilterByVisibility(VisibilityScopeModule)
		assert.Equal(t, &Realm{Name: "test"}, filtered)
	})
	t.Run("keeps higher visibility", func(t *testing.T) {
		realm := &Realm{
			Name: "test",
			Modules: []*Module{{
				Name:  "test",
				Decls: []Decl{&Verb{Name: "test", Visibility: VisibilityScopeModule}},
			}},
		}
		filtered := realm.FilterByVisibility(VisibilityScopeNone)
		assert.Equal(t, realm, filtered)
	})
	t.Run("includes referred elements", func(t *testing.T) {
		realm := &Realm{
			Name: "test",
			Modules: []*Module{{
				Name:  "other",
				Decls: []Decl{&Verb{Name: "test2", Visibility: VisibilityScopeNone}},
			}, {
				Name: "test",
				Decls: []Decl{&Verb{
					Name:       "test",
					Visibility: VisibilityScopeModule,
					Metadata: []Metadata{
						&MetadataCalls{Calls: []*Ref{{Module: "other", Name: "test2"}}},
					},
				}},
			}},
		}
		filtered := realm.FilterByVisibility(VisibilityScopeModule)
		assert.Equal(t, realm, filtered)
	})
}
