from enum import Enum

TOP_CAM = "top-cam"
FRONT_CAM = "front-cam"

EMULSIFIER_VIAL = 4

# before, during, or after stable milk in the bowl

PIXEL_FONT = "../resources/fonts/monogram.ttf"
CS_FONT = "../resources/fonts/Comic-Sans-MS.ttf"
MAIN_FONT = "../resources/fonts/DejaVuSerifCondensed-Italic.ttf"


class CanvasStatus(Enum):
    BEFORE = 1
    DURING = 2
    AFTER = 3

# Scenes for defining composition of top and front cams


class Scene(Enum):
    UNDEFINED = 1
    FRONT_ONLY = 2
    DUAL = 3
    TOP_ONLY = 4


class Format(Enum):
    UNDEFINED = 1
    LANDSCAPE = 2
    PORTRAIT = 3
