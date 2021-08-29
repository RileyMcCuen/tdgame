# tdgame (Work In Progress)

tdgame is a work in progress effort to create a tower defense game in the same style as the old flash tower defense game.
My motivations are to get better at various aspects of go, designing complex systems, and to build a game in a dying genre (flash games are dead).

It uses the ebiten game library, which is very lightweight and allows me to experiment with all of the systems I want to design from the ground up.
I also use gg for handling some primitive drawing methods and some image manipulation, but this may be removed in the future if I try to do 
everything through the shaders support in ebiten.

The approaches and methods used to make this game run are not always the best or most performant, but allow me to experiment with various ideas.
Testing may be added if some functionality settles down, but for now anything might change so testing would largely be duplicated wasted effort.

Some of the concepts that have been tinkered with so far include:
- declaritive asset management -- inspiration clearly taken from kubernetes
  - allows less technical contributors to be able to create new assets and add them to the project without any actual code
  - creates a clear order of operations for asset loading
  - allows the game to handle assets differently based on hints provided in the asset declarations
- chunked collision detection based on "tiles" that make up the map
- clean and simple game interface -- the update method used by the ebiten engine makes for very easy state tracking
  - could be leveraged by an ai to play the game as the enemy spawner or player -- eventually plan to make a q learning agent
  
  Currently no ui exist to actually play the game. Eventually there will be, but still evaluating whether to use a premade ui library like ebitenui
  or to make the necessary functionality needed to operate the game.
 Minimum ui components:
    - textbox
    - button
    - card-carousel
    
I will continue to work on the game as time allows, but it will likely not be playable with a full ui for quite some time!
One other contributor and I also make all of the assets so that can take a lot of time as well given that neither of us have much background in art or design.
