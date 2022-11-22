import paho.mqtt.client as mqtt
from pycommon.const import HOST
import yaml
import paho.mqtt.publish as mqttpub
import paho.mqtt.subscribe as mqttsub

CLIENT_ID="py-interfaces"
DEBUG = False

GOTO_NODE_TOPIC = "mega/req/goto-node"
GOTO_TOPIC = "mega/req/goto-xy"
GOTO_RESP_TOPIC = "mega/resp/goto-xy"
DISPENSE_TOPIC = "mega/req/dispense"
DISPENSE_RESP_TOPIC = "mega/resp/dispense"
COLLECT_TOPIC = "mega/req/collect"
SLEEP_TOPIC = "mega/req/sleep"
WAKE_TOPIC = "mega/req/wake"
UNCALIBRATE_TOPIC = "mega/req/uncalibrate"
OPEN_DRAIN_TOPIC = "mega/req/open-drain"
CLOSE_DRAIN_TOPIC = "mega/req/close-drain"

# debug print
def debug(msg):
    if DEBUG:
        print("[MQTT]", msg)

def pub(topic, pl):
    mqttpub.single(topic, payload=pl, hostname=HOST, port=1883, client_id=CLIENT_ID, keepalive=2)

def sub(topic):
    return mqttsub.simple(topic, hostname=HOST, port=1883, client_id=CLIENT_ID, keepalive=2)


def goto_xy(x, y):
    pl = "{:.3f},{:.3f}".format(x, y)

    debug("writing goto_xy payload '{}'".format(pl))
    pub(GOTO_TOPIC, pl)

    debug("wrote goto_xy payload.")# Listening for response...")
    #! commented out because there's no timeout supported, so it hangs if
    #! there's no client responding
    # resp = sub(GOTO_RESP_TOPIC)
    # debug("got goto_xy response payload '{}'".format(resp.payload))

def dispense(ul):
    debug("writing dispense payload '{}'".format(ul))
    pub(DISPENSE_TOPIC, ul)

    debug("wrote dispense payload")#. Listening for response...")
    # resp = sub(DISPENSE_RESP_TOPIC)
    # debug("got dispense response payload '{}'".format(resp.payload))

def collect(pos):
    debug("writing collect payload '{}'".format(pos))
    pub(COLLECT_TOPIC, pos)

    debug("wrote collect payload")

def sleep():
    pub(SLEEP_TOPIC, "")

def wake():
    pub(WAKE_TOPIC, "")

def uncalibrate():
    pub(UNCALIBRATE_TOPIC, "")

def set_drain(b: bool):
    if b:
        pub(OPEN_DRAIN_TOPIC, "")
    else:
        pub(CLOSE_DRAIN_TOPIC, "")
    
def goto_node(node):
    pub(GOTO_NODE_TOPIC, node)