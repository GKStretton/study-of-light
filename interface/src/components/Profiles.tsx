import { Button, Grid, Typography } from "@mui/material";
import "./Profiles.css";
import { useSystemVialProfiles, useVialProfiles } from "../util/hooks";
import { SystemVialConfiguration, VialProfile, VialProfileCollection } from "../machinepb/machine";
import { KV_KEY_ALL_VIAL_PROFILES, KV_KEY_SYSTEM_VIAL_PROFILES, TOPIC_KV_GET } from "../topics_backend/topics_backend";
import { useContext } from "react";
import MqttContext from "../util/mqttContext";

// I still don't understand the type/value distinction that gives rise to this
// const VialProfileCollectionMethods = VialProfileCollection;

export default function Profiles() {
  const [vialProfiles, setVialProfiles] = useVialProfiles();
  const [systemVialProfiles, setSystemVialProfiles] = useSystemVialProfiles();

  // Experimentation
  // const [vialProfiles, setVialProfiles] = useKeyValueStore<typeof VialProfileCollectionMethods>(
  //   KV_KEY_ALL_VIAL_PROFILES,
  //   VialProfileCollection
  // );

  const { client: c } = useContext(MqttContext);

  const requestProfiles = () => {
    if (!c || !c.connected) {
      console.error(`cannot request profiles, client not connected: ${c}`);
      return;
    }
    c?.publish(TOPIC_KV_GET + KV_KEY_ALL_VIAL_PROFILES, "");
    c?.publish(TOPIC_KV_GET + KV_KEY_SYSTEM_VIAL_PROFILES, "");
  };

  // testing
  const setProfiles = () => {
    const msg = vialProfiles ?? VialProfileCollection.create();
    if (!msg.profiles[0]) {
      msg.profiles[0] = VialProfile.create();
    }
    msg.profiles[0].description += "a";
    console.log("setting", msg);
    setVialProfiles(msg);
  };

  const setSystemProfiles = () => {
    const msg = systemVialProfiles ?? SystemVialConfiguration.create();
    msg.vials[0]++;
    setSystemVialProfiles(msg);
  };

  return (
    <>
      <Typography variant="h3">vialProfiles</Typography>
      <Typography variant="body1" sx={{ color: "red" }}>
        Edit file at /mnt/md0/light-stores/kv/vial-profiles, then refresh
      </Typography>
      <Button onClick={setProfiles}>Test</Button>
      <textarea id="profiles" readOnly value={JSON.stringify(vialProfiles, null, 2)}></textarea>
      <Typography variant="h3">systemVialProfiles</Typography>
      <Typography variant="body1" sx={{ color: "red" }}>
        Edit file at /mnt/md0/light-stores/kv/system-vial-profiles, then refresh
      </Typography>
      <Button onClick={setSystemProfiles}>Test</Button>
      <textarea id="systemProfiles" readOnly value={JSON.stringify(systemVialProfiles, null, 2)}></textarea>
      <Button onClick={requestProfiles}>Force request profiles</Button>
    </>
  );
}
