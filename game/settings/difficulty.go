package settings

type (
	Difficulty struct {
		Name                                   string
		EnemyHealthMultiplier                  float32
		WraithMeleeDamage                      float32
		ExplosionMaxDamage, ExplosionMinDamage float32
	}
)

var (
	Difficulties = [...]Difficulty{
		{
			Name:                  "Oh no!",
			EnemyHealthMultiplier: 0.5,
			WraithMeleeDamage:     8.0,
			ExplosionMaxDamage:    25.0,
			ExplosionMinDamage:    0.0,
		},
		{
			Name:                  "I am prepared.",
			EnemyHealthMultiplier: 0.75,
			WraithMeleeDamage:     12.0,
			ExplosionMaxDamage:    35.0,
			ExplosionMinDamage:    5.0,
		},
		{
			Name:                  "Don't hold back.",
			EnemyHealthMultiplier: 1.0,
			WraithMeleeDamage:     15.0,
			ExplosionMaxDamage:    50.0,
			ExplosionMinDamage:    10.0,
		},
	}
)

func CurrDifficulty() Difficulty {
	return Difficulties[Current.DifficultyIndex]
}
