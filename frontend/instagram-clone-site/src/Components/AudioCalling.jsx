import React, {
  useEffect,
  useRef,
  useState,
  useImperativeHandle,
  forwardRef,
} from "react";
import { useOutletContext } from "react-router-dom";

const AudioCalling = forwardRef(
  (
    {
      passedCurrentUserID,
      subscribeNotifications: propSubscribeNotifications,
      sendNotifications: propSendNotifications,
    },
    ref,
  ) => {
    // states
    const [callState, setCallState] = useState("idle"); //["idle","calling","hangup","active"."incoming"]
    const [incomingCallFrom, setIncomingCallFrom] = useState(null); //* for tracking who is calling
    const [incomingOfferSdp, setIncomingOfferSdp] = useState(null);
    const [seconds, setSeconds] = useState(0);
    const [finalDuration, setFinalDuration] = useState("");

    console.log("/audio_calling state:", callState);

    // webRtc & network refs
    const peerConnection = useRef(null); //* for peer rtc connection (gateway)
    const localStream = useRef(null); //* storing active client's media's audio track continuous stream
    const targetPeerID = useRef(null); //* the one current user is connected to <- sent by the backend answer
    const iceCandidatesQueue = useRef([]); //* adding ice-candidate means routes are known and ready for p2p connection -> added when reciever is yet to set remote desc -> untill then candidates are queued
    const remoteAudioRef = useRef(null);

    // * Stun/turn servers for locating client's internet location
    // * client phone asks what's my location on internet -> findable -> once found -> ice candidate is completed -> ready to connect and exchange audio streams from p.c
    const rtcConfig = {
      iceTransportPolicy: "all",
      iceServers: [
        // One robust master STUN server handles 99% of P2P lookups
        { urls: "stun:stun.l.google.com:19302" },

        // Streamlined TURN relay channels favoring stable ports
        {
          urls: "turn:openrelay.metered.ca:80",
          username: "openrelayproject",
          credential: "openrelayproject",
        },
        {
          urls: "turn:openrelay.metered.ca:443",
          username: "openrelayproject",
          credential: "openrelayproject",
        },
        {
          urls: "turn:openrelay.metered.ca:443?transport=tcp", // Forces reliable TCP over flaky UDP
          username: "openrelayproject",
          credential: "openrelayproject",
        },
      ],
    };

    // ** peer connection & webRTC flow **//

    // 1. offer request is made -> it says ' hey i am available to connect, i have opened my media streams and added to the peer connection {aka p.c} would you like to connect?
    // 2. answer response is made to respond to the 'offer' request which -> says ' yes,i have opened my media streams too and added stream to the established peer connection,i'm ready to connect.
    // > inner-context => when offer is made it carries offer payload which -> carries information what request is being sent and what info will be recieved by the reciever for sharing his media stream
    // same way, when answer reponds, it carries answer payload which -> carries information about what answer will be sent and recieved and decoded by reciever{ctx - caller} to let them know - we are sharing media stream and ready to recieve media stream
    // > also, offer/answer sdp payload carries information like 'what is being sent and how', not like it helps determine the paths or how peers will be connection through networking
    // its more like, offer sdp payload says ' i support these codecs,settings,opened my stream and added to peer connection, i want to recieve audio'
    // then whem answer with attached answer sdp responds it says ' i got your request, i have also opened my media stream {audiotrack}, i'm available and want to recieve audio too'
    // 3. The actual networking is done by ice-candidates -> this determines route and figures out how 'peers would be connected by routing',
    //  STUN server sole job is to get client's Internet ip address and candidate litreally takes this information to route peers in that address where connection is made

    // $ Flow becomes
    // offer sdp -> says 'what it will be sharing to p.c and what it wants'
    // answer sdp -> responds ' what it will aslo be sharing and what it wants'
    // ice candidate -> get both parties ip's by help of STUN servers and figures out address/route to connect peers for media streams
    // so ice-candidates are exchanges after sdp's agrees

    // @ Local and remote things
    //1. when offer is created -> it fires setLocalDesc(that offer is made)-> then parallely browser starts gathering ice-candidates -> which fires onicecandidates multitimes\
    // but yet, offer(localdesc) is not setremote, so reciever cannot add them (to ice candidates) as remote is not set -> gets queued for later proccessing
    //2. then when answer responds -> set remoteDesc(answer is made) -> creates a answer sdp -> sets localDsec(answer)
    // now once remoteDesc(answer) is set -> reciever can process those queued candidates which got queued cause reciever had yet to set remote desc
    // now=> you might be wondering what is this setting local/remoteDescription -> these are just locally (if localDesc) telling offer is made,shared stream is ready to set remote which will be recieved by the reciever
    // that's why when answer setRemoteDesc -> desc simply sets answer which means -> it is too ready to share media streams
    // desc are set to notify peers as they are ready to share media streams

    // $ local v/s remote flow
    // 1. Setting local descrption -> like what local stream setting description ,codecs and information needs to be sent -> what it needs to proccess incoming audio on other side
    // 2. Setting remote description -> means like setting remote description of incoming streams -> this is my description of stream,codecs and incoming remotely as audio information -> what it needs to let browser side knows
    // that it tells ' what would be incoming and ready for decoding incoming streams'  -> it also supports these codecs and configuration -> so from other side knows what is incoming and what audio to decode
    // 3. then when local decription of stream and remote are set, both knows about each other streams and codec configs -> fires on track to access and play the incoming audio

    // ** end ** //

    const outletContext = useOutletContext();
    const sendNotifications =
      propSendNotifications || outletContext?.sendNotifications;
    const subscribeNotifications =
      propSubscribeNotifications || outletContext?.subscribeNotifications;

    const logTrackDetails = (prefix, track) => {
      if (!track) {
        console.log(`[WebRTC Telemetry] ${prefix} - No track provided`);
        return;
      }
      try {
        const settings = track.getSettings ? track.getSettings() : {};
        const capabilities = track.getCapabilities
          ? track.getCapabilities()
          : {};
        console.log(`[WebRTC Telemetry] ${prefix} Track Details:`, {
          id: track.id,
          kind: track.kind,
          label: track.label,
          enabled: track.enabled,
          muted: track.muted,
          readyState: track.readyState,
          settings: JSON.stringify(settings),
          capabilities: JSON.stringify(capabilities),
        });
        // ! note: do NOT set onmute/onunmute here - caller sets them in ontrack handler
        // ! overwriting them here would kill the audio restart logic
        track.onended = () => {
          console.log(`[WebRTC Telemetry] ${prefix} Track ENDED:`, track.id);
        };
      } catch (err) {
        console.error(
          `[WebRTC Telemetry] Error logging track details for ${prefix}:`,
          err,
        );
      }
    };

    const setupPeerConnectionListeners = (pc) => {
      console.log(
        "[WebRTC Telemetry] Setting up listeners on RTCPeerConnection",
      );

      pc.oniceconnectionstatechange = () => {
        const state = pc.iceConnectionState;
        console.log(
          `[WebRTC Telemetry] ICE Connection State changed to: ${state}`,
        );
        if (state === "failed" || state === "disconnected") {
          console.warn(
            "[WebRTC Telemetry] ICE Connection is in failed/disconnected state.",
          );
        }
        // ! belt-and-suspenders: when ICE confirms a live path, re-trigger audio element play()
        // ! in case ontrack fired before ICE and the initial play() hit silence
        if (state === "connected" || state === "completed") {
          const el = remoteAudioRef.current;
          if (el && el.srcObject) {
            console.log(
              "[WebRTC Telemetry] ICE connected - re-triggering audio element play()",
            );
            el.muted = false;
            el.volume = 1.0;
            el.play().catch((err) =>
              console.error(
                "[WebRTC Telemetry] Play on ICE connect failed:",
                err,
              ),
            );
          }
        }
      };

      pc.onsignalingstatechange = () => {
        console.log(
          `[WebRTC Telemetry] Signaling State changed to: ${pc.signalingState}`,
        );
      };

      pc.onconnectionstatechange = () => {
        console.log(
          `[WebRTC Telemetry] Connection State changed to: ${pc.connectionState}`,
        );
      };

      pc.onicegatheringstatechange = () => {
        console.log(
          `[WebRTC Telemetry] ICE Gathering State changed to: ${pc.iceGatheringState}`,
        );
      };

      pc.onicecandidateerror = (e) => {
        console.error("[WebRTC Telemetry] ICE Candidate Error:", {
          errorCode: e.errorCode,
          errorText: e.errorText,
          url: e.url,
        });
      };
    };

    // Cleanup/Hangup handling
    const handleHangup = () => {
      // ! whem hanged up safely cleam dead rtc connection or open streams n mics
      if (peerConnection.current) {
        peerConnection.current.ontrack = null;
        peerConnection.current.onremovetrack = null;
        peerConnection.current.onicecandidate = null;
        peerConnection.current.oniceconnectionstatechange = null;
        peerConnection.current.onsignalingstatechange = null;

        // closing the peer rtc connection
        peerConnection.current.close();
        peerConnection.current = null;
      }

      // since local stream carries the current caller media devices
      if (localStream.current) {
        localStream.current?.getTracks().forEach((track) => {
          console.log("stopping all media audio track streams...");
          track.stop(); // stopping all mic streams
          console.log("successfully stopped all media audio track streams🛑.");
        });
        localStream.current = null;
        console.log("local media stream is set to null🛑");
      }

      // tearing down the hidden audio served element for tracks
      const remoteAudioElement = remoteAudioRef.current;
      if (remoteAudioElement) {
        remoteAudioElement.srcObject = null;
        console.log("teared down srcObj audio source; srcObj is null now🛑");
      }

      iceCandidatesQueue.current = [];
      console.log("teared down ice candidate queues;empty queue array now");
      setCallState("hangup");
      console.log("call state is now 'hangup'");
      console.log(
        "user has been disconnected from the call; all the rtc connection have been safely remved.",
      );
    };

    const triggerHangup = (notifyPeer = true) => {
      if (notifyPeer && sendNotifications) {
        const peerId = targetPeerID.current || incomingCallFrom;
        if (peerId) {
          sendNotifications({
            sender_id: Number(passedCurrentUserID),
            reciever_id: Number(peerId),
            type: "hangup",
          });
        }
      }
      handleHangup();
    };

    // Expose IntialiseCalling trigger to the parent component
    useImperativeHandle(ref, () => ({
      initialiseCalling: (recieverID) => {
        IntialiseCalling(recieverID);
      },
    }));

    //! Incoming call accept handler - invokes when call is accpeted
    const acceptCall = async () => {
      try {
        console.log(" Accepting incoming call offer from:", incomingCallFrom);
        if (!incomingOfferSdp) {
          console.warn(
            "No incomingOfferSdp available!;unable to accept call as sdp payload is not ready to notify other side about sending and recieving stream info❌",
          );
          return;
        }

        // findme
        //  opening p.c connection from stun config -> onicecandidate to add candidate for networking ->
        // ->  ontrack for sourcing recieved remote stream to the client -> setting up remote desc -> queuped up candidates proccessing ->
        // -> local media audio stream opening -> adding stream to p.c connection -> create ans sdp ->
        // -> setting up local desc for proccessing incoming stream -> sending ans payload

        //**  correct flow
        // 1. rtc conn setup
        // parralelly adding and setting up ice candidates for routing and networking for p2p calling
        // 2. handle incoming streaming by labeling and locking incoming remote stream description with setRemoteDes
        // queued candidates
        // 3. open client's media stream
        // 4. Grab local mic stream and add to the p.c
        // 5. create ans sdp
        // 6. construct payload
        // 7. send via ws connection
        //**  **//

        // peer connection is stored in current state <- created by new RTCPeerConnection(passingInIceStunServersConfig)
        peerConnection.current = new RTCPeerConnection(rtcConfig); // just remember everything is stored in current state by the use of ref
        console.log(
          "peer connection is opened successfully✅;configured in with stun servers config.",
        );
        setupPeerConnectionListeners(peerConnection.current);

        // before setting up connection,adding this to candidate for ipLookup by stun servers and candidate path matching for interconnection

        // * configuring and adding icecandidates/queuing up
        peerConnection.current.onicecandidate = (e) => {
          //**  sending payload of audio_type payload of context - e.candidate => if e.candidate exists in peerConn. here --
          if (e.candidate && sendNotifications) {
            console.log(
              "ice candidate is generated locally for this client✅; started gathering information for this client's connection.",
            );

            const peerId = targetPeerID.current || incomingCallFrom;

            sendNotifications({
              sender_id: Number(passedCurrentUserID),
              reciever_id: Number(peerId),
              type: "ice-candidate",
              audio_payload_only: e.candidate,
            });
          } else {
            console.log(
              "upon onicecandidate invocation; it has successfully gathered ice candidate information and ready for connection✅... ",
            );
          }
        };

        // * streaming remote already added media tracks streams
        peerConnection.current.ontrack = (e) => {
          const track = e.track; // always use the track directly, not e.streams[0] which can be muted/empty on first fire
          // on track property gives us remotely recieved stream <- in streams array being the 0th as first element being the recieved stream
          const remoteAudioStream = e.streams[0];

          console.log(
            "remote audio stream which was added to p.c, it's details:",
            {
              streamsCount: e.streams.length,
              trackKind: track.kind,
              trackId: track.id,
              trackMuted: track.muted,
              trackReadyState: track.readyState,
            },
          );
          logTrackDetails("Remote (Accept)", track);
          console.log(
            "successfully accessing already added media streams in the p.c✅;ready to play⏳...",
          );

          let remoteHiddenAudioElement = remoteAudioRef.current;

          // if element exists and cause it would be hidden, sourcing the recieved stream from peerConnection rtc to source in from e.streams[at0thPlace]
          if (remoteHiddenAudioElement) {
            // ! setting {src} of this hidden el to play this audio track steam from the candidate
            remoteHiddenAudioElement.srcObject = remoteAudioStream;
            remoteHiddenAudioElement.autoplay = true;
            remoteHiddenAudioElement.playsInline = true;
            remoteHiddenAudioElement.muted = false; // explicitly unmute - browser can default to muted for autoplay policy

            // ! critical: ontrack fires BEFORE ICE connects - track starts muted, play() hits silence.
            // ! bind onunmute to restart playback the moment ICE connects and audio starts flowing
            track.onunmute = () => {
              console.log(
                "Track found UNMUTED (ICE connected, streams flowing) - trying to unmute and play again...",
              );
              remoteHiddenAudioElement.play().catch((err) => {
                console.error(
                  "failed to play media strea,;caught error when unmute:",
                  err,
                );
              });
            };

            remoteHiddenAudioElement
              .play()
              .then(() => {
                console.log(
                  "succesfully played remote mic stream;already added to the p.c✅.",
                );
                console.log("active audio element state (Receiver):", {
                  muted: remoteAudioRef.current.muted,
                  volume: remoteAudioRef.current.volume,
                  srcObject: remoteAudioRef.current.srcObject
                    ? "has stream"
                    : "NO STREAM",
                  paused: remoteAudioRef.current.paused,
                });
              })
              .catch((err) => {
                console.error(
                  "failed to play remote mic stream; which was added to the p.c :",
                  err,
                );
              });
          }
        };

        // setting session description to be from recieved 'offer' payload from ws connection
        console.log(
          "Setting remote description (Offer); so other side would know about incoming stream and configs ...",
        );

        // * setting up remote description for -> decoding and acknowleding incoming remote stream
        await peerConnection.current.setRemoteDescription(
          new RTCSessionDescription(incomingOfferSdp),
        );
        console.log(
          "successfully setted up remote description of stream; other side would know what is being sent and its configuration for playing incoming stream✅",
        );

        // * queued candidates processing when remote description from other side are not set
        if (iceCandidatesQueue.current.length > 0) {
          console.log(
            `[WebRTC Accept] Processing ${iceCandidatesQueue.current.length} queued ICE candidates on accept`,
          );
          for (const candidate of iceCandidatesQueue.current) {
            // ! skip null/end-of-gathering candidates
            if (
              !candidate ||
              (!candidate.candidate &&
                candidate.sdpMid == null &&
                candidate.sdpMLineIndex == null)
            ) {
              console.log(
                "could not find any ice-candidates;candidate information for tracking and connecting peer is unknown ❌",
              );
              continue;
            }
            try {
              console.log(
                "trying to queuing up ice candidates untill other side remote description is not set up⏳... ",
                candidate,
              );
              await peerConnection.current.addIceCandidate(
                new RTCIceCandidate(candidate),
              );
              console.log(
                "successfully added up ice candidate✅;one side is ready to recieve incoming streams",
              );
            } catch (err) {
              console.error("failed to add queued ice candidate:", err);
            }
          }
          iceCandidatesQueue.current = [];
          console.log("ica candidate queue is set to be empty array.");
        }

        // get his mics and all and store in localStream as streamingMic
        console.log("trying to request media stream of client⏳...");

        // * mic stream grabbing
        localStream.current = await window.navigator.mediaDevices.getUserMedia({
          audio: true,
          video: false,
        });

        console.log(
          "successfully recieved media stream✅; mic audio stream is now available to be added to the peer connection⏳...",
        );

        localStream.current.getTracks().forEach((track, i) => {
          logTrackDetails(`Local [${i}]`, track);
        });

        //* adding mic stream to the pc connection
        localStream.current.getTracks().forEach((track) => {
          console.log(
            "grabbing client already accessed mic stream, trackID:",
            track.id,
          );
          peerConnection.current.addTrack(track, localStream.current);
          console.log(
            "successfully added media audio stream track to the p.c✅;",
            track.id,
          );
        });

        // now sending answer payload to the caller with sdp block and audio_payload, this time sending 'answer' payload,
        console.log(" Creating answer repsonse...");

        // * creating answer sdp
        const createdAnsSdpPayload = await peerConnection.current.createAnswer({
          offerToReceiveAudio: true,
        }); //sdp answer audio payload, sending to caller with type "answer" for connection
        console.log(
          "answer sdp is created successfully✅; Setting local description to decode incoming stream...;other side must have setted up remote desc so local description would know about decoding and playing incoming media stream...",
        );

        // * setting local description for -> sending information about sending stream configs
        await peerConnection.current.setLocalDescription(createdAnsSdpPayload);
        console.log(
          "local description is setted up successfully;ready to recieve streams to play✅;other side must have setted up remote descriotion for local description to know incoming stream to handle and play",
        );

        //  every offer/ans/candidate_req payload is sent to ws, attaching audio_payload based off context of outbound requests

        // * sending answer sdp attached payload to the ws
        if (sendNotifications) {
          console.log(
            "Sending answer notification via WS to targetPeerID:",
            targetPeerID.current,
          );
          const peerId = targetPeerID.current || incomingCallFrom;
          sendNotifications({
            sender_id: Number(passedCurrentUserID),
            reciever_id: Number(peerId), // senderId - id of the user whose 'offer' request was intercepted and in response 'answer' is created for him wiht ans audio_payload attached inside
            type: "answer", // sending ans payload to the ws with type being "answer" -> by unmarshaling it would know what is incoming audio_payload and what needs to be published & shipped to consume by the reciever
            audio_payload_only: createdAnsSdpPayload,
          });
        } else {
          console.warn(
            "sendNotifications functionality is not available to send ANSWER",
          );
        }

        setCallState("active");
        console.log("call state is now set to 'active'");
      } catch (error) {
        console.error("failed to accept incoming call:", error);
        triggerHangup(true);
      }
    }; //..accept call

    // ws connection is already there, just subcribing to it and retreving information

    // ** All hub's redirected payloads for negotiations are recieved and handled here
    useEffect(() => {
      if (!subscribeNotifications) return;

      const unsubscribe = subscribeNotifications(async (audioPayload) => {
        //** interceptor recieved the payload on onmessage
        if (!audioPayload) {
          console.error("could not get any  sdp request payload.");
          return;
        }

        switch (
          audioPayload.type //& when call 'offer' is recieved <- for cal connection => reciever get that request with peerID being the senderID as sender is one who is sending call request
        ) {
          case "offer": {
            // when offer sdp is sent -> what it is being sent/what it expects (media audio track stream)
            console.log(` offer request is recieved ✅`, {
              type: audioPayload.type,
              sdpType: audioPayload.audio_payload_only?.type,
            });
            targetPeerID.current = audioPayload.sender_id; // sender_id is what supplied by hub from reciever attaching -> the senderID ;reciever sends the sender id as peerID
            setIncomingCallFrom(audioPayload.sender_id);
            setIncomingOfferSdp(audioPayload.audio_payload_only);
            setCallState("incoming");
            console.log("offer request is successfully decoded.", {
              sentBY_peerID: audioPayload.sender_id,
            });
            break;
          }
          case "answer": {
            console.log(`answer request is recieved ✅`, {
              type: audioPayload.type,
              sdpType: audioPayload.audio_payload_only?.type,
            });
            // &when call is either approved or not <- 'answering' call => reciever gets 'answer''s audio_payload
            if (!peerConnection.current) {
              console.warn(
                "answer request is successfully recieved;but peerConnection.current does not hold any p.c❌",
              );
              break;
            }
            // ! guard: only set remote description if we're in the right signaling state
            // ! if 3 handlers fire for this same message, 2nd and 3rd would throw "wrong state" without this check
            if (peerConnection.current.signalingState !== "have-local-offer") {
              console.warn(
                `Skipping setRemoteDescription(answer) - wrong signalingState: ${peerConnection.current.signalingState}`,
              );
              break;
            }
            try {
              console.log(
                "setting up remote description of my media stream codec configs on reciever side⏳",
              );
              console.log(
                "for letting other side know what my stream codec configs would be and what i want...",
              );
              await peerConnection.current.setRemoteDescription(
                // * setting description on other side that -> what i would send and want, caller side is ready for incoming streams now✅
                new RTCSessionDescription(audioPayload.audio_payload_only),
              );
              console.log(
                "successfully setted up remote description of outgoing stream codecs configs information on reciever side✅",
              );
              console.log(
                "other side is now locked in,ready to recieve incoming media streams;successfully configured to decode incoming stream to play.",
              );
              setCallState("active");

              // Process queued candidates
              if (iceCandidatesQueue.current.length > 0) {
                console.log(
                  `queuing ice-candidates,cause only local description is setup,but unknown of what other side would be sending🤔❓;`,
                );
                console.log(
                  "reciever has not configured remote description yet👎,incoming media stream config is unknowm, queuing up candidates🛳️",
                );
                for (const candidate of iceCandidatesQueue.current) {
                  // ! skip null/end-of-gathering candidates
                  if (
                    !candidate ||
                    (!candidate.candidate &&
                      candidate.sdpMid == null &&
                      candidate.sdpMLineIndex == null)
                  ) {
                    console.log(
                      "could not find any ice-candidates;candidate information for tracking and connecting peer is unknown❌",
                    );
                    console.log(
                      "null payload;failed to decode peer ip address which is needed for connecting them❌",
                    );
                    continue;
                  }
                  try {
                    console.log(
                      "trying to queuing up ice candidates untill other side remote description is not set up⏳... ",
                      candidate,
                    );
                    await peerConnection.current.addIceCandidate(
                      new RTCIceCandidate(candidate),
                    );
                    console.log(
                      "successfully queued up ice candidate✅;one side is ready to recieve incoming streams",
                    );
                    console.log(
                      "local description of streams and audio codecs configs are successfully setted up;waiting for other side to set remote description⏳",
                    );
                  } catch (err) {
                    console.error("failed to add queued ICE candidate❌:", err);
                  }
                }
                iceCandidatesQueue.current = [];
              }
            } catch (err) {
              console.error(
                "failed to proccess incoming answer sdp payload attached on incoming audio payload❌",
                err,
              );
              // * local description must be set both sides for setting up config for incoming stream and then it must be setted up  remote description for other side
              // * this way -> local ( for incoming media audio track stream recognition amd remote for sharing config what would be sent,no local could know and decode incoming media stream)
            }
            break;
          }
          case "ice-candidate": {
            //& for connecting up peers on peer connection for shared audio stream listening

            // fires parallely when local and remote desc are setup and offer is accepted and answer responded
            const candidate = audioPayload.audio_payload_only;
            console.log(
              ` successfully recieved ice-candidate payload from sender: ${audioPayload.sender_id}`,
              candidate,
            );
            console.log(
              "ice candidate is ready to route sender to peer connection location site after getting theirs actual internet ip address location from stun servers⏳...",
            );

            // ! guard: null candidate = end-of-gathering signal from the browser, not a real candidate
            // ! constructing RTCIceCandidate from it throws "sdpMid and sdpMLineIndex are both null"
            if (
              !candidate ||
              // these information let the stun servers to find internet location of the peers.
              // if these are not available -> ice candidate would not be able to get ip address of peer and to connect them p2p for audio stream connection
              (!candidate.candidate &&
                candidate.sdpMid == null &&
                candidate.sdpMLineIndex == null)
            ) {
              console.log(
                "could not find any information in candidate sdp payload which would have been helpful to connect peers❌;",
              );
              console.log(
                "null payload;failed to decode peer ip address which is needed for connecting them❌",
              );
              break;
            }

            if (
              peerConnection.current && //p.c exists
              peerConnection.current.remoteDescription // p.c also does have setted up remote desc -> other side is now able to decode and play incoming media stream
            ) {
              try {
                console.log(
                  "p.c is available and also client has setted up remote description of stream successfully on the other side✅",
                );
                console.log(
                  "ice candidates trying to connect peers for media streaming connection⏳...",
                );
                await peerConnection.current.addIceCandidate(
                  new RTCIceCandidate(candidate),
                );
                console.log(
                  "successfully added ice candidate✅;client is ready for listening media stream from other side✅",
                );
              } catch (err) {
                console.error(
                  "despite of having active p.c and remote description being settedup;failed to add ICE candidate:❌",
                  err,
                );
              }
            } else {
              // Queue candidate for later
              iceCandidatesQueue.current.push(candidate);
              console.log(
                "queuing up ice-candidates;other side has not configured and set its remote description of the stream❌",
                candidate,
              );
            }
            break;
          }
          case "hangup": {
            console.log(
              `Received HANGUP request from sender: ${audioPayload.sender_id}`,
            );
            //& handle hangup state <- when peerID becomes unavailable -> hangup the call // asking reciever client to hang up the calling connection and media devices as peer has been disconnected
            handleHangup();
            break;
          }
        }
      });

      return () => {
        if (unsubscribe) {
          console.log("unsubcribing audio calling ws connection"); // cleanup ws connection from dead loose ends -> unsubcribe happens when ws connection is lost or gracefully shut downed
          unsubscribe();
        }
      };
    }, [subscribeNotifications, passedCurrentUserID]); //only re render calling service if current userID get alernated unexpectedly

    // Duration Timer management
    useEffect(() => {
      let interval = null;
      if (callState === "active") {
        setSeconds(0);
        setFinalDuration("");
        interval = setInterval(() => {
          setSeconds((prev) => prev + 1);
        }, 1000);
      } else if (callState === "hangup") {
        const mins = Math.floor(seconds / 60)
          .toString()
          .padStart(2, "0");
        const secs = (seconds % 60).toString().padStart(2, "0");
        setFinalDuration(`${mins}:${secs}`);

        // Reset back to idle state after 3 seconds
        const timeout = setTimeout(() => {
          setCallState("idle");
          setSeconds(0);
        }, 3500);

        return () => clearTimeout(timeout);
      } else if (callState === "idle") {
        setSeconds(0);
      }

      return () => {
        if (interval) clearInterval(interval);
      };
    }, [callState]);

    // fires up calling and sending offer to the reciever
    const IntialiseCalling = async (recieverID) => {
      console.log("ref check at start:", remoteAudioRef.current);
      // 1. sending a constructed payload via ws connection to let reader's publisher publish the payload to the exchange ->
      // 2. as always consumer keep chekcing for the incoming delivery in the exchange, if there is any delivery stamped for the reciever with type "offer" ->
      // 3. redirects the constructed payload with audio_payload to the hub's audio chan which ->
      // 4. redirects the incoming payload with attached peer{callingPeer~partner} to the reciever which <- then it's on reciever to either answer or decline

      // we can store anything on the set ref variable's current state <- keeps audioPayload/conn intact
      targetPeerID.current = recieverID; // so setting the target user to be someone for whom this fnct would be invoked to send 'offer'
      try {
        //! Calling flow
        // current state would store essential things
        console.log(
          "initializing firing up of offer sdp request to the reciever:",
          recieverID,
        );
        console.log(
          "offer request is made✅;media streams are yet to be opened and added in p.c and yet to ask other side - if they would like to connect⏳...",
        );
        setCallState("calling");

        // ! pre-activate the remote audio element with the user click gesture to bypass autoplay policy blocks
        const el = remoteAudioRef.current;
        if (el) {
          el.muted = false;
          el.volume = 1.0;
          el.play().catch((err) =>
            console.log(
              "[WebRTC Call] Pre-play gesture activation:",
              err.message,
            ),
          );
        }

        // 1. grab mic/permissions first
        // setting localStream <- senderSide permission for audio access
        console.log("trying to request media stream of client⏳...");
        localStream.current = await navigator.mediaDevices.getUserMedia({
          // mediaDevices are readyOnly props <- read what client has provided to the browser
          audio: true,
          video: false,
        });
        console.log(
          "successfully recieved media stream✅; mic audio stream is now available to be added to the peer connection⏳...",
        );
        const checktrack = localStream.current?.getTracks()[0];
        console.log("testing client stream configurations", {
          hasStreamEnabled: checktrack?.enabled,
          hasStreamMuted: checktrack?.muted,
          streamState: checktrack?.readyState,
        });
        console.log(
          "successfully grabbed media stream of client✅; stream is ready to be added to p.c⏳...",
        );
        localStream.current.getTracks().forEach((track, i) => {
          logTrackDetails(`Local [${i}]`, track); //!todo -  need to check this
        });

        // 2. initialize rtcConnection
        // setting peerConnection to store rtcConnection for audio connection
        peerConnection.current = new RTCPeerConnection(rtcConfig); // passing in stun servers to open peerConnection,which -> is store in it's current state

        console.log(
          "peer connection is opened successfully✅;configured in with stun servers config.",
        );
        setupPeerConnectionListeners(peerConnection.current); //!todo - need to check this later too

        // 3. attach~connect the permitted microphone to reciever side {peer}
        localStream.current.getTracks().forEach((track) => {
          console.log(
            "grabbing client already accessed mic stream, trackID:",
            track.id,
          );
          peerConnection.current.addTrack(track, localStream.current);
          // todo - later we can add logic to track how many streams were added in pc to know if something failed, then p.c would not have got that track stream
          console.log(
            "successfully added media audio stream track to the p.c✅;",
            track.id,
          );
        });

        peerConnection.current.onicecandidate = (e) => {
          // ice candidates are then created to -> fired on onicecandidates -> to create candidates for connection

          // since up till now -> both peer would have been connected to the peerConn built from rtcConnection
          if (e.candidate && sendNotifications) {
            // if they exists, in rtcConnection clients exists as candidates
            // sending 'ice-candidate' type of payload to the ws connection so that to be published and consumed by the reciever with peerID being the calling partner who had made 'offer' request

            console.log(
              "ice candidate is generated locally for this client✅; started gathering information for this client's connection.",
            );

            // if ws.send is available
            const peerId = targetPeerID.current || incomingCallFrom;
            const payload = {
              sender_id: Number(passedCurrentUserID),
              reciever_id: Number(peerId), // stored in its current state, we can use ref to store things dynamically throughout the application
              type: "ice-candidate",
              audio_payload_only: e.candidate,
            };

            sendNotifications(payload); // this would send this type of payload to the client's reader which would be consumed with attached peerID as senderID
            // carries informatio about ice candidate sdp payload attached for connection -> if this is sent successfully and not null -> client is ready for connection
            console.log(peerConnection.current?.getSenders(), "sender"); //todo - need to look into these
          } else {
            console.log(
              "upon onicecandidate invocation; it has successfully gathered ice candidate information and ready for connection✅... ",
            );
          }
        };
        // streaming incoming stream from peerRtcConnection
        peerConnection.current.ontrack = (e) => {
          console.log("ref check in ontrack:", remoteAudioRef.current);
          // **all stored/added media streams are available in this propery

          const track = e.track; // always use the track directly, not e.streams[0] which can be muted/empty on first fire
          // on track property gives us remotely recieved stream <- in streams array being the 0th as first element being the recieved stream
          const remoteAudioStream = e.streams[0];

          //** */ each client recieves other side added streams in p.c

          console.log(
            "remote audio stream which was added to p.c, it's details:",
            {
              streamsCount: e.streams.length,
              trackKind: track.kind,
              trackId: track.id,
              trackMuted: track.muted,
              trackReadyState: track.readyState,
            },
          );

          console.log(
            "successfully accessing already added media streams in the p.c✅;ready to play⏳...",
          );
          logTrackDetails("Remote (Call)", track);
          const remoteHiddenAudioElement = remoteAudioRef.current; // ohhh, this would be played in hidden side but hearble track

          // if element exists and cause it would be hidden, sourcing the recieved stream from peerConnection rtc to source in from e.streams[at0thPlace]

          // ** responsible for playing streams which -> are stored in the p.c
          if (remoteHiddenAudioElement) {
            // if element exists
            // ! setting {src} of this hidden el to play this audio track steam from the candidate
            remoteHiddenAudioElement.srcObject = remoteAudioStream; //* sourcing audio from media stream added in p.c
            remoteHiddenAudioElement.autoplay = true;
            remoteHiddenAudioElement.playsInline = true;
            remoteHiddenAudioElement.muted = false; //* added explicit muted state off
            // ! critical: ontrack fires BEFORE ICE connects - track starts muted, play() hits silence.
            // ! bind onunmute to restart playback the moment ICE connects and audio starts flowing

            // todo - before custom logging,need context and knowledge
            track.onunmute = () => {
              console.log(
                "[WebRTC Call] Track UNMUTED (ICE connected, audio flowing) - restarting play()",
              );
              remoteHiddenAudioElement.play().catch((err) => {
                console.error(
                  "[WebRTC Call] Audio playback failed on unmute:",
                  err,
                );
              });
            };

            console.log("[WebRTC Call] Playing remote audio...");
            remoteHiddenAudioElement
              .play()
              .then(() => {
                console.log("[WebRTC Call] Playback started successfully.");
                console.log("active audio element state (Sender):", {
                  muted: remoteAudioRef.current.muted,
                  volume: remoteAudioRef.current.volume,
                  srcObject: remoteAudioRef.current.srcObject
                    ? "has stream"
                    : "NO STREAM",
                  paused: remoteAudioRef.current.paused,
                });
              })
              .catch((err) => {
                console.error("[WebRTC Call] Audio playback failed:", err);
              });
          }
        };

        // 6. if both parties are okay, create a sdp offer,send to ws handler
        console.log(
          "calling btn clicked; firing a offer sdp attached request ⏳...",
        );
        const offerRequest = await peerConnection.current.createOffer({
          offerToReceiveAudio: true,
        }); // starts a remote rtc connection to the peer, explicitly requesting bidirectional audio
        console.log(
          "offer spd is created successfully✅;yet to set local description of the stream.",
        );
        await peerConnection.current.setLocalDescription(offerRequest); //* sets the session description in peerConn to be this offer request between candidates
        console.log(
          "local description is successfully setted up✅;ready for recieving incoming media stream unless remotedesc is not configured/set by the other side. ",
        );

        if (sendNotifications) {
          // if ws.send is available
          // @ note -> every audio_payload_only is generated from peerConn methods -> all we did is checking the flow and sending candidate req with e.candidate from the peer connection to send that tyoe of payload
          // @ - like this, also attaching payload "offer" created from same peerConnection n <=- every needy thing is created from peerConnection which holds all the things for setting up calling connection
          // @ - which the peerConnection is created from rtc
          const peerId = targetPeerID.current || incomingCallFrom;
          const payload = {
            sender_id: Number(passedCurrentUserID),
            reciever_id: Number(peerId), // stored in its current state, we can use ref to store things dynamically throughout the application
            type: "offer",
            audio_payload_only: offerRequest,
          };

          sendNotifications(payload);
          console.log("offer request is made successfully✅", {
            reciever_id: payload.reciever_id,
          });
        } else {
          console.warn(
            "could not able to send offer request;sendNotification functionality is not available❌",
          );
        }
      } catch (err) {
        console.error("failed to initiate calling📞❌", err);
        triggerHangup(true);
      }
    };

    const getFormattedTime = () => {
      const mins = Math.floor(seconds / 60)
        .toString()
        .padStart(2, "0");
      const secs = (seconds % 60).toString().padStart(2, "0");
      return `${mins}:${secs}`;
    };

    const getStatusText = () => {
      switch (callState) {
        case "calling":
          return "Calling...";
        case "incoming":
          return "Incoming Call";
        case "active":
          return "Connected";
        case "hangup":
          return "Call Ended";
        default:
          return "";
      }
    };

    const peerLabel = targetPeerID.current || incomingCallFrom || "User";

    return (
      <>
        <audio
          ref={remoteAudioRef}
          id="remote-hidden-audio-element"
          autoPlay
          playsInline
          style={{ display: "none" }}
        />
        {callState !== "idle" && (
          <div className="audiocall-overlay">
            <div className="audiocall-card">
              {/* Pulsating avatar placeholder */}
              <div
                className={`audiocall-avatar ${
                  callState === "calling" || callState === "active"
                    ? "pulse-calling"
                    : callState === "incoming"
                      ? "pulse-incoming"
                      : ""
                }`}
              >
                {String(peerLabel).substring(0, 2).toUpperCase()}
              </div>

              <h3 className="audiocall-name">User #{peerLabel}</h3>

              <p
                className={`audiocall-status ${callState === "active" ? "is-active" : "is-pending"}`}
              >
                {getStatusText()}
              </p>

              {/* Dynamic call timer */}
              {callState === "active" && (
                <div className="audiocall-timer">{getFormattedTime()}</div>
              )}

              {/* Final Talk duration display */}
              {callState === "hangup" && finalDuration && (
                <div className="audiocall-duration">
                  Call Duration:{" "}
                  <span className="audiocall-duration-value">
                    {finalDuration}
                  </span>
                </div>
              )}

              {/* Action Controls */}
              <div className="audiocall-actions">
                {callState === "incoming" ? (
                  <>
                    {/* Accept button */}
                    <button
                      onClick={acceptCall}
                      className="audiocall-btn audiocall-btn-accept"
                    >
                      📞
                    </button>
                    {/* Decline button */}
                    <button
                      onClick={() => triggerHangup(true)}
                      className="audiocall-btn audiocall-btn-decline"
                    >
                      ✖
                    </button>
                  </>
                ) : (
                  callState !== "hangup" && (
                    /* Hangup action button */
                    <button
                      onClick={() => triggerHangup(true)}
                      className="audiocall-btn audiocall-btn-decline"
                    >
                      🛑
                    </button>
                  )
                )}
              </div>
            </div>
          </div>
        )}
      </>
    );
  },
);

AudioCalling.displayName = "AudioCalling";
export default AudioCalling;
