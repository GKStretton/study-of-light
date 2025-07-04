import { useContext, useState, useEffect, useCallback, useRef } from 'react';
import './ConfigPage.css';
import MqttContext from '../util/mqttContext';
import { Typography, Button, FormControl, InputLabel, Select, MenuItem, Checkbox, FormControlLabel, Paper } from '@mui/material';
import VideoPlayer from './VideoPlayer';
import { TOPIC_KV_GET, TOPIC_KV_SET, TOPIC_KV_GET_RESP, TOPIC_KV_SET_RESP, TOPIC_TRIGGER_DSLR, TOPIC_TRIGGER_DSLR_RESP } from '../topics_backend/topics_backend';

interface CropConfig {
  left_abs: number;
  right_abs: number;
  top_abs: number;
  bottom_abs: number;
  left_rel: number;
  right_rel: number;
  top_rel: number;
  bottom_rel: number;
}

const CAMERA_CHOICES = {
  'top-cam': { id: '1', name: 'Top Camera', configKey: 'crop_top-cam', streamUrl: 'ws://milkelangelo:8889/top-cam/' },
  'front-cam': { id: '2', name: 'Front Camera', configKey: 'crop_front-cam', streamUrl: 'ws://milkelangelo:8889/front-cam/' },
  'dslr': { id: '3', name: 'DSLR', configKey: 'crop_dslr', streamUrl: '' }
};

function ConfigPage() {
  const { client: c, messages } = useContext(MqttContext);
  const [selectedCamera, setSelectedCamera] = useState<keyof typeof CAMERA_CHOICES>('top-cam');
  const [cropConfig, setCropConfig] = useState<CropConfig>({
    left_abs: 0, right_abs: 100, top_abs: 0, bottom_abs: 100,
    left_rel: 0, right_rel: 0, top_rel: 0, bottom_rel: 0
  });
  const [showOverlay, setShowOverlay] = useState(true);
  const [videoDimensions, setVideoDimensions] = useState({ width: 0, height: 0 });
  const [dslrImageUrl, setDslrImageUrl] = useState<string>(`http://milkelangelo:8089/get-dslr-preview?t=${Date.now()}`);
  const [saveStatus, setSaveStatus] = useState<string>('');
  const [dslrDisplayDimensions, setDslrDisplayDimensions] = useState({ width: 0, height: 0 });
  const dslrImageRef = useRef<HTMLImageElement>(null);

  // Subscribe to response topics
  useEffect(() => {
    if (c && c.connected) {
      const configKey = CAMERA_CHOICES[selectedCamera].configKey;
      c.subscribe(TOPIC_KV_GET_RESP + configKey);
      c.subscribe(TOPIC_KV_SET_RESP + configKey);
      c.subscribe(TOPIC_TRIGGER_DSLR_RESP);
    }
  }, [c, selectedCamera]);

  // Load crop config when camera selection changes
  useEffect(() => {
    if (c && c.connected) {
      const configKey = CAMERA_CHOICES[selectedCamera].configKey;
      c.publish(TOPIC_KV_GET + configKey, '');
    }
  }, [selectedCamera, c]);

  // Handle config response messages
  useEffect(() => {
    const configKey = CAMERA_CHOICES[selectedCamera].configKey;
    const configMessage = messages[TOPIC_KV_GET_RESP + configKey];
    
    if (configMessage) {
      try {
        const config = JSON.parse(configMessage.toString());
        if (config) {
          setCropConfig(config);
          console.log('Loaded crop config:', config);
        }
      } catch (e) {
        console.log('No existing config found, using defaults');
      }
    }
  }, [messages, selectedCamera]);

  // Handle save response messages
  useEffect(() => {
    const configKey = CAMERA_CHOICES[selectedCamera].configKey;
    const saveMessage = messages[TOPIC_KV_SET_RESP + configKey];
    
    if (saveMessage) {
      setSaveStatus('Saved successfully!');
      setTimeout(() => setSaveStatus(''), 3000);
      console.log('Crop config saved successfully');
    }
  }, [messages, selectedCamera]);

  // Handle DSLR response messages
  useEffect(() => {
    const dslrMessage = messages[TOPIC_TRIGGER_DSLR_RESP];
    
    if (dslrMessage) {
      const response = dslrMessage.toString();
      if (response === 'success') {
        // Wait 5 seconds before fetching the image to allow processing
        setTimeout(() => {
          setDslrImageUrl(`http://milkelangelo:8089/get-dslr-preview?t=${Date.now()}`);
        }, 5000);
      }
    }
  }, [messages]);

  // Track DSLR image display dimensions
  useEffect(() => {
    const updateDslrDisplayDimensions = () => {
      if (dslrImageRef.current) {
        setDslrDisplayDimensions({
          width: dslrImageRef.current.clientWidth,
          height: dslrImageRef.current.clientHeight
        });
      }
    };

    const resizeObserver = new ResizeObserver(() => {
      updateDslrDisplayDimensions();
    });

    if (dslrImageRef.current) {
      updateDslrDisplayDimensions();
      resizeObserver.observe(dslrImageRef.current);
    }

    return () => {
      if (dslrImageRef.current) {
        resizeObserver.unobserve(dslrImageRef.current);
      }
    };
  }, [dslrImageUrl, selectedCamera]);

  // Handle DSLR capture
  const triggerDslrCapture = useCallback(() => {
    if (c && c.connected && selectedCamera === 'dslr') {
      c.publish(TOPIC_TRIGGER_DSLR, '');
    }
  }, [c, selectedCamera]);

  // Save crop config
  const saveCropConfig = useCallback(() => {
    if (c && c.connected) {
      const configKey = CAMERA_CHOICES[selectedCamera].configKey;
      c.publish(TOPIC_KV_SET + configKey, JSON.stringify(cropConfig));
      setSaveStatus('Saving...');
      
      // Set timeout for save operation
      setTimeout(() => {
        if (saveStatus === 'Saving...') {
          setSaveStatus('Save timeout');
          setTimeout(() => setSaveStatus(''), 3000);
        }
      }, 5000);
    }
  }, [c, selectedCamera, cropConfig, saveStatus]);

  // Common crop config update logic
  const updateCropConfig = useCallback((x: number, y: number, isShiftKey: boolean) => {
    if (isShiftKey) {
      // Set bottom-right corner
      let newRight = x;
      let newBottom = y;

      // For non-front cameras, maintain square aspect ratio
      if (selectedCamera !== 'front-cam') {
        const mag = Math.abs(newRight - cropConfig.left_abs);
        newBottom = cropConfig.top_abs + mag;
      }

      // Constrain to bounds first
      newRight = Math.min(videoDimensions.width, Math.max(cropConfig.left_abs + 2, newRight));
      newBottom = Math.min(videoDimensions.height, Math.max(cropConfig.top_abs + 2, newBottom));

      // Then ensure dimensions are even for x264 compatibility
      if ((newRight - cropConfig.left_abs) % 2 === 1) newRight -= 1;
      if ((newBottom - cropConfig.top_abs) % 2 === 1) newBottom -= 1;

      setCropConfig(prev => ({
        ...prev,
        right_abs: newRight,
        bottom_abs: newBottom
      }));
    } else {
      // Set top-left corner
      let newLeft = x;
      let newTop = y;
      let newRight = cropConfig.right_abs;
      let newBottom = cropConfig.bottom_abs;

      // For non-front cameras, maintain square aspect ratio
      if (selectedCamera !== 'front-cam') {
        const mag = Math.abs(newRight - newLeft);
        newBottom = newTop + mag;
      }

      // Constrain to bounds first
      newLeft = Math.max(0, Math.min(newLeft, cropConfig.right_abs - 2));
      newTop = Math.max(0, Math.min(newTop, cropConfig.bottom_abs - 2));
      newRight = Math.min(videoDimensions.width, Math.max(newLeft + 2, newRight));
      newBottom = Math.min(videoDimensions.height, Math.max(newTop + 2, newBottom));

      // Then ensure dimensions are even for x264 compatibility
      if ((newRight - newLeft) % 2 === 1) newRight -= 1;
      if ((newBottom - newTop) % 2 === 1) newBottom -= 1;

      setCropConfig(prev => ({
        ...prev,
        left_abs: newLeft,
        top_abs: newTop,
        right_abs: newRight,
        bottom_abs: newBottom
      }));
    }
  }, [videoDimensions, cropConfig, selectedCamera]);

  // Handle mouse events for crop selection
  const handleVideoClick = useCallback((e: React.MouseEvent<HTMLVideoElement>) => {
    // Get actual video dimensions from the video element
    const videoElement = e.currentTarget;
    const actualWidth = videoElement.videoWidth || videoElement.clientWidth;
    const actualHeight = videoElement.videoHeight || videoElement.clientHeight;
    
    // Update videoDimensions if not set or different
    if (videoDimensions.width !== actualWidth || videoDimensions.height !== actualHeight) {
      setVideoDimensions({ width: actualWidth, height: actualHeight });
    }

    if (!actualWidth || !actualHeight) return;

    const rect = videoElement.getBoundingClientRect();
    const x = Math.round((e.clientX - rect.left) * (actualWidth / rect.width));
    const y = Math.round((e.clientY - rect.top) * (actualHeight / rect.height));

    updateCropConfig(x, y, e.shiftKey);
  }, [videoDimensions, updateCropConfig]);

  // Handle mouse events for image clicks (DSLR)
  const handleImageClick = useCallback((e: React.MouseEvent<HTMLImageElement>) => {
    if (!videoDimensions.width || !videoDimensions.height) return;

    const rect = e.currentTarget.getBoundingClientRect();
    const x = Math.round((e.clientX - rect.left) * (videoDimensions.width / rect.width));
    const y = Math.round((e.clientY - rect.top) * (videoDimensions.height / rect.height));

    updateCropConfig(x, y, e.shiftKey);
  }, [videoDimensions, updateCropConfig]);

  // Update relative values when absolute values change
  useEffect(() => {
    if (videoDimensions.width && videoDimensions.height) {
      setCropConfig(prev => ({
        ...prev,
        left_rel: prev.left_abs,
        right_rel: videoDimensions.width - prev.right_abs,
        top_rel: prev.top_abs,
        bottom_rel: videoDimensions.height - prev.bottom_abs
      }));
    }
  }, [cropConfig.left_abs, cropConfig.right_abs, cropConfig.top_abs, cropConfig.bottom_abs, videoDimensions]);

  // Render crop overlay
  const renderCropOverlay = useCallback((displayDimensions: { width: number; height: number }) => {
    if (!showOverlay || !displayDimensions.width || !displayDimensions.height || !videoDimensions.width || !videoDimensions.height) return null;

    // Scale from actual video coordinates to display coordinates
    const scaleX = displayDimensions.width / videoDimensions.width;
    const scaleY = displayDimensions.height / videoDimensions.height;

    return (
      <div className="crop-overlay">
        <div
          className="crop-rectangle"
          style={{
            left: cropConfig.left_abs * scaleX,
            top: cropConfig.top_abs * scaleY,
            width: (cropConfig.right_abs - cropConfig.left_abs) * scaleX,
            height: (cropConfig.bottom_abs - cropConfig.top_abs) * scaleY,
          }}
        />
      </div>
    );
  }, [showOverlay, cropConfig, videoDimensions]);

  return (
    <div style={{padding: "1rem"}}>
      <br/>
      <Typography variant="h5">Config Page</Typography>
      
      <Paper style={{ padding: '1rem', marginTop: '1rem' }}>
        <Typography variant="h6" gutterBottom>Crop Configuration Tool</Typography>
        
        <div className="crop-instructions">
          <Typography variant="body2" gutterBottom>
            <strong>Instructions:</strong>
          </Typography>
          <ul>
            <li>Select a camera from the dropdown</li>
            <li>Click on the video/image to set the top-left corner of the crop area</li>
            <li>Hold Shift and click to set the bottom-right corner</li>
            <li>For top-cam and DSLR, the crop area will maintain a square aspect ratio</li>
            <li>Click "Save Crop Config" to store your settings</li>
          </ul>
        </div>
        
        <div className="crop-controls">
          <FormControl style={{ minWidth: 200 }}>
            <InputLabel>Camera</InputLabel>
            <Select
              value={selectedCamera}
              onChange={(e) => setSelectedCamera(e.target.value as keyof typeof CAMERA_CHOICES)}
            >
              {Object.entries(CAMERA_CHOICES).map(([key, camera]) => (
                <MenuItem key={key} value={key}>{camera.name}</MenuItem>
              ))}
            </Select>
          </FormControl>
          
          <FormControlLabel
            control={
              <Checkbox
                checked={showOverlay}
                onChange={(e) => setShowOverlay(e.target.checked)}
              />
            }
            label="Show Overlay"
          />
          
          {selectedCamera === 'dslr' && (
            <Button variant="contained" onClick={triggerDslrCapture}>
              Capture DSLR Image
            </Button>
          )}
          
          <Button variant="contained" color="primary" onClick={saveCropConfig}>
            Save Crop Config
          </Button>
          
          {saveStatus && (
            <Typography variant="body2" color={saveStatus.includes('success') ? 'primary' : 'error'}>
              {saveStatus}
            </Typography>
          )}
        </div>

        <div className="crop-info">
          <Typography variant="body2" color="textSecondary">
            Click to set top-left corner, Shift+Click to set bottom-right corner
          </Typography>
          <div className="coordinates">
            Crop: ({cropConfig.left_abs}, {cropConfig.top_abs}) to ({cropConfig.right_abs}, {cropConfig.bottom_abs})
            {' '}Size: {cropConfig.right_abs - cropConfig.left_abs} × {cropConfig.bottom_abs - cropConfig.top_abs}
          </div>
        </div>

        <div className="crop-container" style={{ maxWidth: '800px', margin: '0 auto' }}>
          {selectedCamera === 'dslr' ? (
            dslrImageUrl ? (
              <div style={{ position: 'relative' }}>
                <img
                  ref={dslrImageRef}
                  className="crop-image"
                  src={dslrImageUrl}
                  alt="DSLR Preview"
                  style={{
                    transform: 'rotate(180deg)',
                    maxWidth: '100%',
                    height: 'auto'
                  }}
                  onClick={handleImageClick}
                  onLoad={(e) => {
                    const img = e.target as HTMLImageElement;
                    setVideoDimensions({ width: img.naturalWidth, height: img.naturalHeight });
                  }}
                />
                {renderCropOverlay(dslrDisplayDimensions)}
              </div>
            ) : (
              <div className="dslr-placeholder">
                <Typography>Click "Capture DSLR Image" to load preview</Typography>
              </div>
            )
          ) : (
            <VideoPlayer
              url={CAMERA_CHOICES[selectedCamera].streamUrl}
              name={`crop-${selectedCamera}`}
              handleClick={handleVideoClick}
              renderOverlay={renderCropOverlay}
              onVideoLoad={(videoElement: HTMLVideoElement) => {
                // Set video dimensions when video loads
                if (videoElement.videoWidth && videoElement.videoHeight) {
                  setVideoDimensions({
                    width: videoElement.videoWidth,
                    height: videoElement.videoHeight
                  });
                }
              }}
            />
          )}
        </div>
      </Paper>
    </div>
  )
}

export default ConfigPage;