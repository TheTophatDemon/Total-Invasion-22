package settings

import (
	"fmt"
)

type (
	Difficulty struct {
		// Index into the difficulties array
		Index int
		// Multiplies the starting health of all enemies
		EnemyHealthMultiplier float32
		// Number of seconds after seeing you before an enemy starts to pursue
		EnemyWakeTime float32
		// Amount of damage each hit from the Wraith's melee attack deals
		WraithMeleeDamage float32
		// Damage range based on distance to the source of an explosion
		ExplosionMaxDamage, ExplosionMinDamage float32
		// Number of seconds between shots form the Dummkopf's mouth
		DummkopfShootSpeed float32
		// Multiplies the speed of non-player projectiles
		ProjectileSpeedFactor float32
	}
)

var (
	Difficulties = [...]Difficulty{
		{
			Index:                 0,
			EnemyHealthMultiplier: 0.5,
			EnemyWakeTime:         0.5,
			WraithMeleeDamage:     8.0,
			ExplosionMaxDamage:    25.0,
			ExplosionMinDamage:    0.0,
			DummkopfShootSpeed:    0.3,
			ProjectileSpeedFactor: 0.7,
		},
		{
			Index:                 1,
			EnemyWakeTime:         0.5,
			EnemyHealthMultiplier: 0.75,
			WraithMeleeDamage:     12.0,
			ExplosionMaxDamage:    35.0,
			ExplosionMinDamage:    5.0,
			DummkopfShootSpeed:    0.2,
			ProjectileSpeedFactor: 0.9,
		},
		{
			Index:                 2,
			EnemyWakeTime:         0.3,
			EnemyHealthMultiplier: 1.0,
			WraithMeleeDamage:     15.0,
			ExplosionMaxDamage:    50.0,
			ExplosionMinDamage:    10.0,
			DummkopfShootSpeed:    0.15,
			ProjectileSpeedFactor: 1.0,
		},
	}
)

func (diff Difficulty) String() string {
	return Localize(fmt.Sprintf("difficulty%d", diff.Index))
}

func CurrDifficulty() Difficulty {
	return Difficulties[Current.DifficultyIndex]
}
