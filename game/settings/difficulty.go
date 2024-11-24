package settings

type (
	Difficulty struct {
		Name                                   string
		WraithMeleeDamage                      float32
		ExplosionMaxDamage, ExplosionMinDamage float32
	}
)

var (
	Difficulties = [...]Difficulty{
		{
			Name:               "Oh no!",
			WraithMeleeDamage:  8.0,
			ExplosionMaxDamage: 25.0,
			ExplosionMinDamage: 0.0,
		},
		{
			Name:               "I am prepared.",
			WraithMeleeDamage:  12.0,
			ExplosionMaxDamage: 35.0,
			ExplosionMinDamage: 5.0,
		},
		{
			Name:               "Don't hold back.",
			WraithMeleeDamage:  15.0,
			ExplosionMaxDamage: 50.0,
			ExplosionMinDamage: 10.0,
		},
	}
)

func CurrDifficulty() Difficulty {
	return Difficulties[Current.DifficultyIndex]
}
