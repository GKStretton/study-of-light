import React, { useState, useEffect, useContext } from 'react';
import { WebRTCReceiver } from '../util/WebRTCReceiver';
import { Button, Box, Typography } from '@mui/material';

interface VideoPlayerProps {
  url: string;
  name: string;
  handleClick?: (e: React.MouseEvent<HTMLVideoElement>) => void;
  renderOverlay?: (videoDimensions: { width: number; height: number }) => React.ReactNode;
  onVideoLoad?: (videoElement: HTMLVideoElement) => void;
}

const VideoPlayer: React.FC<VideoPlayerProps> = ({ url, name, handleClick, renderOverlay, onVideoLoad }: VideoPlayerProps) => {
  const [displayDimensions, setDisplayDimensions] = useState<{ width: number; height: number }>({ width: 0, height: 0 });

  useEffect(() => {
    const receiver = new WebRTCReceiver(url, name);
  }, [url, name]);

  useEffect(() => {
    const videoElement = document.getElementById(name) as HTMLVideoElement;

    const updateDisplayDimensions = () => {
      if (videoElement) {
        // Use display dimensions for overlay positioning
        setDisplayDimensions({
          width: videoElement.clientWidth,
          height: videoElement.clientHeight
        });
      }
    };

    const handleVideoLoad = () => {
      if (videoElement && onVideoLoad) {
        onVideoLoad(videoElement);
      }
    };

    const resizeObserver = new ResizeObserver(() => {
      updateDisplayDimensions();
    });

    if (videoElement) {
      updateDisplayDimensions();
      resizeObserver.observe(videoElement);
      
      // Add event listeners for video loading
      videoElement.addEventListener('loadedmetadata', handleVideoLoad);
      videoElement.addEventListener('loadeddata', handleVideoLoad);
    }

    return () => {
      if (videoElement) {
        resizeObserver.unobserve(videoElement);
        videoElement.removeEventListener('loadedmetadata', handleVideoLoad);
        videoElement.removeEventListener('loadeddata', handleVideoLoad);
      }
    };
  }, [name, onVideoLoad]);


  return (
    <Box display="flex" flexDirection="column" alignItems="center">
      <div style={{ position: 'relative' }}>
        <video
          id={name}
          muted
          controls={false}
          onClick={handleClick}
          autoPlay
          playsInline
          style={{
            width: '100%',
            height: '100%',
            border: '1px solid black',
          }}
        ></video>
        {renderOverlay && renderOverlay(displayDimensions)}
      </div>
      <Box display="flex" flexDirection="row" alignItems="center">
        {/* <Typography>{JSON.stringify(displayDimensions)}</Typography> */}
      </Box>
    </Box>
  );
};

export default VideoPlayer;
