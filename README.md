# Docker Swarm Visualiser

More intended for users of docker swarm than administrators, this connects to a swarm and shows information and allows some interaction.

## Todo

* [ ] Parallelise and background the context-switch based tab refreshes
    * [ ] Use the "Active" section of the tool bar to show when things are happening in the background
* [ ] Don't add buttons for things you're not authorised to interact with
    * [ ] Add a command to show/ hide services/ storage/ secrets not in your VLAD auth
* [ ] Catch the context switch to limit it to a small return window, cancelling any actions if the docker version returns a failure.