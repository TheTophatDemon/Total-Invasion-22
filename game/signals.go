package game

type (
	MapChangeSignal struct {
		NextMapPath     string
		GiveAmmo        Ammo
		GiveArmor       ArmorType
		ArmorAmount     float32
		EquippedWeapons [WeaponCount]bool
	}
	ResumeGameSignal    struct{}
	TeleportationSignal struct{}
)
