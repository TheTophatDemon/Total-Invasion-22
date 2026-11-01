# TODO - Development
- Add texture properties to make props consistent in size
- Need ammo indicator for weapon wheel
- Some kind of feedback when shooting weapon is blocked?
- Make chickens whackable with sign
- Consider making doors not close automatically (it could surprise players who are trying to flee)
- Make eyeballs transparent
- Switch save system to picking one at the start of the game
- Chicken cannon secondary attack
- Gamepad support (for Steam Deck)
- Bugs
  - Fix jittering in corners
  - Jittering while walking across non-solid props sometimes??
  - When changing resolution after throwing a sickle, the sickle will be in the wrong animation.
- Have world iterators return iter.Seq2 (it might not actually heap allocate)
- Replace GLFW bindings with purego bindings to reduce CGO overhead
- Optimize the memory usage of map loading to avoid GC stutters
- Optimize memory usage of particle effects
- Optimize pauses by pre loading textures for entities

# TODO - Art
- Modify loading screen graphic to be more on-model
- Exit screen
- Make checkpoint effects less harsh
- Make item flashing less harsh / add setting
- Player hit noise
- Dialog Cutscene
  - Song
- Demo end screen
- Re-record enemy voices

# TODO - Levels
- E1M5
  - Animate the clowns
  - Add crowd cheering sound
- E1M6
  - Polish
  - Song

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
