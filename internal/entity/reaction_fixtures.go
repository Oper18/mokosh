package entity

type ReactionMap map[string]Reaction

func (m ReactionMap) Get(name string) Reaction {
	if result, ok := m[name]; ok {
		return result
	}
	return Reaction{}
}

func (m ReactionMap) Pointer(name string) *Reaction {
	if result, ok := m[name]; ok {
		return &result
	}
	return &Reaction{}
}

var reactionComment = "Great shot!"

var ReactionFixtures = ReactionMap{
	"PhotoAliceLove": {
		ID:        1,
		PhotoUID:  PhotoFixtures.Get("Photo01").PhotoUID,
		UserUID:   UserFixtures.Get("alice").UserUID,
		Emoji:     "❤️",
		CreatedAt: *TimeStamp(),
	},
	"PhotoBobComment": {
		ID:        2,
		PhotoUID:  PhotoFixtures.Get("Photo01").PhotoUID,
		UserUID:   UserFixtures.Pointer("bob").UserUID,
		Comment:   &reactionComment,
		CreatedAt: *TimeStamp(),
	},
}

// CreateReactionFixtures inserts known entities into the database for testing.
func CreateReactionFixtures() {
	for _, entity := range ReactionFixtures {
		Db().Create(&entity)
	}
}
