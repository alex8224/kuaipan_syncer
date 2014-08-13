package consts

import (
    inotify "code.google.com/p/go.exp/inotify"
)

const (
    QUIT = 100
    IN_CREATE_DIR = inotify.IN_ISDIR | inotify.IN_CREATE
    IN_DELETE_DIR = inotify.IN_ISDIR | inotify.IN_DELETE
    MASK = IN_CREATE_DIR | IN_DELETE_DIR | inotify.IN_CREATE | inotify.IN_CLOSE_WRITE | inotify.IN_MOVED_TO | inotify.IN_MOVED_FROM | inotify.IN_DELETE
    IN_MOVED_DIR_FROM = inotify.IN_ISDIR | inotify.IN_MOVED_FROM
    IN_MOVED_DIR_TO = inotify.IN_ISDIR | inotify.IN_MOVED_TO
    TYPE_FILE = 200
    TYPE_DIR = 300
)


