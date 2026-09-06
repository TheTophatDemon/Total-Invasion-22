package settings

import (
	"fmt"
)

type (
	Difficulty struct {
		Index                                  int
		EnemyHealthMultiplier                  float32
		WraithMeleeDamage                      float32
		ExplosionMaxDamage, ExplosionMinDamage float32
	}
)

var (
	Difficulties = [...]Difficulty{
		{
			Index:                 0,
			EnemyHealthMultiplier: 0.5,
			WraithMeleeDamage:     8.0,
			ExplosionMaxDamage:    25.0,
			ExplosionMinDamage:    0.0,
		},
		{
			Index:                 1,
			EnemyHealthMultiplier: 0.75,
			WraithMeleeDamage:     12.0,
			ExplosionMaxDamage:    35.0,
			ExplosionMinDamage:    5.0,
		},
		{
			Index:                 2,
			EnemyHealthMultiplier: 1.0,
			WraithMeleeDamage:     15.0,
			ExplosionMaxDamage:    50.0,
			ExplosionMinDamage:    10.0,
		},
		{
			Index:                 3,
			EnemyHealthMultiplier: 1.5,
			WraithMeleeDamage:     20.0,
			ExplosionMaxDamage:    75,
			ExplosionMinDamage:    15.0,
		},
	}
)

func (diff Difficulty) String() string {
	return Localize(fmt.Sprintf("difficulty%d", diff.Index))
}

func CurrDifficulty() Difficulty {
	return Difficulties[Current.DifficultyIndex]
}
