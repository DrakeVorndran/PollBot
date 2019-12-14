<p align="center">
  <img src="gopher-typing.gif" height="250">
</p>

# PollBot
Pollbot is a bot for slack designed to create simple polls based on the reactions of users

[![Go Report Card](https://goreportcard.com/badge/github.com/droxey/goslackit)](https://goreportcard.com/report/github.com/droxey/goslackit) [![Codacy Badge](https://api.codacy.com/project/badge/Grade/7ed40f9f3ecf46709879d5fbac28fd9b)](https://www.codacy.com/app/droxey/goslackit?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=droxey/goslackit&amp;utm_campaign=Badge_Grade)

# [Demo](youtube.com/video/wqmx_vNcjMk)
short youtube video showing the features of pollbot

# Usage
Once PollBot is part of your slack you can create read and delete polls with simple commands listed below

# Commands
Note every command is prefixed with @PollBot so that it knows you are talking to it

* ### Create `<PollName> [Poll Options]`
  Creates a poll with the name `PollName` and asigns reactions to each option you give
* ### Read `<PollName>`
  Prints the current standings of the poll `PollName` in a human readable format with both number of votes and overall percentage
* ### End `<PollName>`
  Starts by printing the standings, then deletes the poll from the database locking the standings at what they currently are
* ### Commands `[Command]`
  not given a command, this will list every command that pollbot understands, with a command this will give a short description of the command as well as the proper usage of the command.
