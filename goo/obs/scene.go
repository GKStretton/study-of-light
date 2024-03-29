package obs

import (
	"fmt"

	"github.com/andreykaipov/goobs/api/requests/scenes"
)

const (
	SCENE_ERROR    = "error"
	SCENE_COMPLETE = "complete"
	SCENE_PAUSED   = "paused"
	SCENE_LIVE     = "live"
)

func setScene(sceneName string) error {
	lock.Lock()
	defer lock.Unlock()
	if c == nil {
		return fmt.Errorf("cannot set obs scene because client is nil")
	}
	_, err := c.Scenes.SetCurrentProgramScene(&scenes.SetCurrentProgramSceneParams{
		SceneName: sceneName,
	})
	if err != nil {
		return fmt.Errorf("failed to set scene: %v", err)
	}
	fmt.Printf("set scene to %s\n", sceneName)
	return nil
}
