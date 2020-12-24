# Generating a dataloader

Use the dataloaden library: https://github.com/vektah/dataloaden
e.g. dataloaden FontLoader uuid.UUID "\*github.com/montfont/sherpa/internal/db/models.Font"

Various modifications need to be made to the generated code.

It doesn't work well with go modules, so you have to run it in the main directory, then move it to this directory and change the package from main to dataloaders.

It also does not import types properly because it seems to assume they're all in the package main. Manually import those.
