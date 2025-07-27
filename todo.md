# TODO
- Eyeball messages
  - Drop some lore in the secret room on E1M2
- Title screen
- Player hit noise
- Chicken cannon secondary attack
- Game Difficulty Settings
- Dialog Cutscene
- Demo end screen
- Settings menu
- Gamepad support (for Steam Deck)

- Fix collision jittering
  - Resolve map collisions first
  - Do multiple iterations on the collision solver
  - Will need to refactor collision callbacks to fire only once
- Fix flickering artifacts on sloped walls
  - Mostly happens on tiles that have some kind of rotation applied to them.
  - Also shows up when spawning on test-teleporters map
  - Could be solved using greedy meshing
- Use texture config for additive rendering and convert to TOML
- Re-record enemy voices
- Make enemies able to hear player behind walls as long as space is connected.

- Port E1M4
  - Textures
    - Satanic fountain
    - Guardrail
  - Song

## Roadmap after Demo release
- Change asset loading to use .zip packages?
- Save states
- Additional Enemies
  - Prisrak
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
  - Sign of Madness
  - Defenestrator (Ep 4)
  - Cluckster Bomb (Ep 4)

## Ideas
- Episode 4: Titan Loose Ends
  - Electric sand
  - Wraiths on motorcycles??
  - Armor that makes game run in slow motion when you are firing your weapon
  - Use graphics based on old prototypes of Total Invasion
