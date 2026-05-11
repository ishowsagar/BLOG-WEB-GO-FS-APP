import { useRef, useState } from "react";

export default function Reels() {
  const videoRefs = useRef([]);
  // Sample video URLs
  const sampleVideos = [
    "https://www.w3schools.com/html/mov_bbb.mp4",
    "https://www.w3schools.com/html/movie.mp4", // replaced with a working video
    "https://filesamples.com/samples/video/mp4/sample_640x360.mp4",
    // Add more URLs if you want more variety
  ];
  // Create an array of 50 videos by repeating sampleVideos
  const videos = Array.from(
    { length: 50 },
    (_, i) => sampleVideos[i % sampleVideos.length]
  );
  const [playing, setPlaying] = useState(Array(videos.length).fill(false));
  // !  will check which index has active node attached --> play it
  //   function handlePlay(id) {
  //     if (videoRefs.current[id]) {
  //       videoRefs.current[id].play();
  //     }
  //   }

  function handleToggle(id) {
    const video = videoRefs.current[id];
    if (video) {
      if (video.paused) {
        video.play();
        setPlaying((prev) => prev.map((p, i) => (i === id ? true : p)));
      } else {
        video.pause();
        setPlaying((prev) => prev.map((p, i) => (i === id ? false : p)));
      }
    }
  }

  const videoElements = videos.map((videoSrc, index) => (
    <div key={index}>
      <video
        ref={(el) => (videoRefs.current[index] = el)}
        loop
        muted
        controls={false}
        className="reel-video"
        onClick={() => handleToggle(index)}
        style={{ cursor: "pointer" }}
        src={videoSrc}
      />
    </div>
  ));
  return (
    <div className="reels_page_container">
      <div className="reels-grid">{videoElements}</div>
    </div>
  );
}
