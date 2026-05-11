export default function Explore() {
  // ! generate an array of empty 100 items/el, wehere each empty el would be an image
  //* this will direclty give array of 100 images src
  const images = Array.from(
    { length: 100 },
    (_, i) => `https://picsum.photos/300/300?random=${i}`
  );

  const randomSizeImage = Math.floor(Math.random() * images.length);

  const image = images.map((imageSrc, index) => (
    <img
      className={`explore-img${
        randomSizeImage === index ? "explore-img-large" : ""
      }`}
      key={index}
      src={imageSrc}
    />
  ));

  return (
    <div className="explore_page_wrapper">
      <div className="explore_page_grid-layout">{image}</div>
    </div>
  );
}
