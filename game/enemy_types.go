package game

type EnemyType uint8

const (
	ENEMY_TYPE_WRAITH EnemyType = iota
	ENEMY_TYPE_FIRE_WRAITH
	ENEMY_TYPE_MOTHER_WRAITH
	ENEMY_TYPE_DUMMKOPF
	ENEMY_TYPE_COUNT
)
