# TODO - Development
- Changing difficulty does not update level layout after loading save
  - Reload the whole level when changing the difficulty setting, preserve equipment
- Punish Me difficulty
  - Disable overhealing
  - Make enemies revive?
  - Allow enemies to open doors
  - Hide menu option until you hover over it
- No ammo indicator for weapon wheel
- Double Grenade Launcher
- Chicken cannon secondary attack
- Fix jittering in corners
- Jittering while walking across non-solid props sometimes??
- Gamepad support (for Steam Deck)
- Have world iterators return iter.Seq2 (it might not actually heap allocate)
- Replace GLFW bindings with purego bindings to reduce CGO overhead
  - Would it be worth it to do the same with tdaudio?
- Optimize the memory usage of map loading to avoid GC stutters
- Optimize memory usage of particle effects
- Optimize pauses by pre loading textures for entities
- When changing resolution after throwing a sickle, the sickle will be in the wrong animation.

# TODO - Art
- Modify loading screen graphic to be more on-model
- Exit screen
- Make checkpoint effects less harsh
- Player hit noise
- Dialog Cutscene
  - Song
- Demo end screen
- Re-record enemy voices

# TODO - Levels
- E1M6
  - Song
  - Double Grenade Launcher

## Roadmap after Demo release
- Change asset loading to use .zip packages?
- Additional Enemies
  - Providence
  - Fundie
  - Banshee
  - Mutant Wraith
  - Mecha Boss
  - Dummkopf Pawn
  - Tophat Demon
  - New enemies for Episode 4
- Additional Weapons
  - Double Grenade Launcher
  - Defenestrator (Ep 4)
  - Cluckster Bomb (Ep 4)

## Ideas
- Episode 4: Titan Loose Ends
  - Electric sand
  - Wraiths on motorcycles??
  - Armor that makes game run in slow motion when you are firing your weapon
  - Use graphics based on old prototypes of Total Invasion
