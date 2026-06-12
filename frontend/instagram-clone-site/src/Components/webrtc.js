//! findme notes
const video_calling_architecture_notes = () => {
  //&  video calling architecture
  // reciever("ans") side flow -> when you are the reciever, you handle remote incoming stream first (setting up remote desc + getting remote track -> play)
  // -> create ans sdp -> then we create local desc to let other side what it would be recieving ( but first we handle recieved stream, them at the end configure what needs to be sent) ->
  // -> at the end we send fuly constructed ans attached payload to other side via ws connection

  //1. intialize web rtc connection

  //2. parallely firing rtc handlers - onicecandidates + ontracks ( fired automatically)
  //adding ice candidates (unless other side remote desc is not set) -> queue up candidates in arr

  const webRtc_handler_notes = {
    // handlers :
    handlers: {
      //&1. onicecandidates
      //SENDER side:
      // onicecandidate fires → gathers ice -> if ('gathered' ice exists) e.candidate exists → send via WS

      // RECEIVER side:
      // candidate arrives via WS
      // → new RTCIceCandidate(candidate)  ← parsing/wrapping
      // → addIceCandidate()               ← adding to pc
      // → browser tests this route for p2p shared stream connection
      onicecandidates: () => {
        //1. onicecandidate -> what to do when e.candidate (browser starts gathering 'ice') (a possible route for networking) is gathered by browser (when it finds where is client located on the internet) and exists
        // but the incoming information of the candidates comes in plain js, -> 'ice' is gathered
        // caller sends this information via ws connection with attached plain 'ice'

        // (recieves plain js candidate sdp in payload ) -> (parses) -> (add that to p.c)
        addIceCandidate: () => {
          // * reciever side, when payload.type-"ice-candidate" is recieved -> adds the ice -> ✅ connected
          //! -> so 'gathered' webRtc recognized ice in payload.audio_payload(e.candidate sdp) is recieved from ws intercepted notification event
          // once,instance of plain js candidate is decoded and parsed gracefully by creating a new instance of webrtc recognized ice candidate
          // webRTc's RTCIceCandidate(rawIce) func takes in raw plai js ice -> decodes it and converts it into webRTC recognized ice for connection
          //! now this -> instance of recognized webRtc candidate is ready to be added into the p.c candidates -> for connection
          // that candidate is added, by addIceCandidate func on p.c
          // e.g pc.current.addIceCandidate("192.168.1.5:54321 udp ~ parsedCandidate instance")
          // then registering starts kicking off
          // once that is done -> ice connected ✅, route is added from one side ✅
        };
      },
      // & 2. ontrack
    },
  };

  // that's why webrtc handlers should be invoked before doing anything but after rtc conn is setted up -> all these things fired instantly

  //3. setting up remoteDesc for incoming stream (handeling incoming first)

  //4. on track to access remotely added media stream in p.connection
  // dom video node access -> source stream there on larger video element

  //5. add ice candidates if that established successfully -> ice exchange for routing and peer connecting to each other for vide call stream

  //6. grab local media stream audio + video (explicitly) configured
  // storing grabbed stream in localstream
  // sourcing local stream into the local video element

  //7. creating answer sdp
  //8. setting up local desc (letting other side know what would be recieved/and what needs to send)

  //9. constructig full payload of type "ans"

  //10. sending via web socket connection
};
