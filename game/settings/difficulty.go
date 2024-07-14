package settings

type (
	Difficulty struct {
		Name              string
		WraithMeleeDamage float32
	}
)

var (
	Difficulties = [...]Difficulty{
		{
			Name:              "Oh no!",
			WraithMeleeDamage: 8.0,
		},
		{
			Name:              "I am prepared.",
			WraithMeleeDamage: 12.0,
		},
		{
			Name:              "Don't hold back.",
			WraithMeleeDamage: 15.0,
		},
	}
)

func CurrDifficulty() Difficulty {
	return Difficulties[Current.DifficultyIndex]
}
