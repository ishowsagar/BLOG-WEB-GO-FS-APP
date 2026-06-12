import React, {
  useState,
  useImperativeHandle,
  forwardRef,
  useRef,
  useEffect,
} from "react";
import { useOutletContext } from "react-router-dom";

const VideoCalling = forwardRef(
  (
    {
      passedCurrentUserID,
      subscribeNotifications: propSubscribeNotifications,
      sendNotifications: propSendNotifications,
    },
    ref,
  ) => {
    // * states - ref states stores in current property
    const [callState, setCallState] = useState("idle"); // ["idle", "active","incoming","hangup"]
    const [isVideoOff, setIsVideoOff] = useState(false);
    const [isMuted, setIsMuted] = useState(false);
    const localstream = useRef(null); // storing local client media stream audio + video
    const remotestream = useRef(null); // remote incoming stream
    const peerConnection = useRef(null); //* for storing peer connection
    const targetPeerID = useRef(null); //* stores targetted peerIDS
    const [incomingOfferSdp, setIncomingOfferSdp] = useState(null); //* for storing incoming offer payload.offer sdp which <- stored in audio_payload_only

    // refs for accessing remote playback elements dynamically and unaffected by re-renders
    const remoteVideoRef = useRef(null);
    const localVideoRef = useRef(null);
    const queueRecieverCandidates = useRef([]); //queuing candidates untill other side has not responded/ have rtc connection & remote setted up

    // accessing outlet passed down props -> value
    const outletcontext = useOutletContext();
    const sendNotifications =
      propSendNotifications || outletcontext?.sendNotifications;
    const subscribeNotifications =
      propSubscribeNotifications || outletcontext?.subscribeNotifications;
    // * video config codecs constraints

    // rtc config
    const rtcConfig = {
      // rtc config - for rtc connection, need STUN servers/ turn for relay back
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

    // func whose sole purpose is to -> do the cleanup only **hangup notification would be sent seperately -> so it ensures it does not cause both to send notification**
    function handleCleanup() {
      console.log("hanging up call...");

      setIsVideoOff(false);
      setIsMuted(false);
      setIncomingOfferSdp(null);
      targetPeerID.current = null;

      // first notify other side too, we are gonna hang up call
      // so other side case executes to do the same ->

      if (peerConnection?.current) {
        peerConnection.current.ontrack = null;
        peerConnection.current.onremovetrack = null;
        peerConnection.current.onicecandidate = null;
        peerConnection.current.oniceconnectionstatechange = null;
        peerConnection.current.onsignalingstatechange = null;

        // closing connection
        peerConnection.current.close();
        peerConnection.current = null;

        console.log(
          "cleared up loose peer connection and closed all the handlers🛑.",
        );
      }

      // stopping opened local streams
      if (localstream.current) {
        // * only do cleanup if they exists -: avoid cleaning up when does not exists
        localstream.current.getTracks().forEach((eachTrackStream) => {
          console.log("stopping all local media streams...");
          eachTrackStream.stop();
          console.log("cleared up and stopped all the opened local streams🛑.");
        });
      }

      localstream.current = null;
      console.log("localstream ref is now null again🛑.");

      // tearing down hidden + displayed streams
      const remoteVideoRefEl = remoteVideoRef.current;
      const localVideoRefEl = localVideoRef.current;

      // stopping playbacks and tearing down the stream sources

      if (remoteVideoRefEl) {
        remoteVideoRefEl.srcObject = null;
        console.log(
          "teared down srcObject video source; srcObject is null now🛑",
        );
      }

      if (localVideoRefEl) {
        localVideoRefEl.srcObject = null;
        console.log(
          "successfully teared down all src obj including this local too; srcObject is null now🛑",
        );
      }
    }

    function handleHangup() {
      if (sendNotifications) {
        sendNotifications({
          sender_id: Number(passedCurrentUserID),
          reciever_id: Number(targetPeerID.current),
          type: "hangup",
          content: "video",
        });
      }

      // call cleanup now -> because then sent hangup will only do the cleanup part
      handleCleanup();
      setCallState("idle");
    }

    // ** loading ws shared connection when components mounts -> payloads are intercepted here

    // bug - needed to use - subscribeNotifications which subcribe the notifications and recieve 'em
    useEffect(() => {
      if (!subscribeNotifications) {
        console.warn("subscribeNotifications function is not available");
        return;
      }
      const unsubscribe = subscribeNotifications(async (p) => {
        //   notifications are subscribed here
        if (!p) {
          console.log(
            "nothing is sent from hub; no payloads have been intercepted for calling requests❌",
          );
          return;
        }

        // Ignore audio calling notifications
        if (
          ["offer", "answer", "ice-candidate", "hangup"].includes(p.type) &&
          p.content !== "video"
        ) {
          console.log("VideoCalling ignoring incoming audio call payload:", p.type);
          return;
        }

        // if payloads exists and are incoming
        switch (p.type) {
          case "offer": {
            console.log("call offer request is recieved⏳📞...");
            console.log(
              `payload of type ${p.type} is recieved successfully✅; sent by user${p.sender_id}`,
            );

            targetPeerID.current = p.sender_id;
            setIncomingOfferSdp(p.audio_payload_only); //* stores incoming offer sdp in state
            setCallState("incoming");

            break;
          } // ..offer
          case "answer": {
            console.log(
              "user${p.sender_id} has sent you answer to the offer...",
            );
            console.log(
              `payload of type ${p.type} is recieved successfully✅; sent by user${p.sender_id}`,
            );

            targetPeerID.current = p.sender_id; // todo - check later

            // ans

            //1.checking peer connection state first
            if (!peerConnection.current) {
              console.warn(
                "answer request is successfully recieved;but peerConnection.current does not hold any p.c❌",
              );
              break;
            }

            // !cause we need to register remote desc - connection must be checked first
            try {
              // as we are in answer, so for answer -> remote desc must be set for incoming stream
              console.log("setting up remote desc...");
              const parsedSessionDescription = new RTCSessionDescription(
                p.audio_payload_only,
              );
              //   ! remember we did not set remote desc while sending offer { as caller } -> cause we had to set localDesc for sending stream in pconn first
              // ! only if payload of type 'answer' is recieved -> we wanna hear the steam and view -> we set that to be true -> remote desc ->set pconn to recognize incoming stream config codecs -> ready for play
              await peerConnection.current.setRemoteDescription(
                parsedSessionDescription,
              );
              console.log("remote description is set up successfully✅");
              setCallState("active"); //when ans is recievev -> means offer req is accepted by the reciever

              //   resolving pushed queued up candidates
              break;
            } catch (err) {
              console.error(err);
            }

            // setting up remote desc -> flow
          } // ..answer
          case "ice-candidate": {
            // ice is sent by the caller, for connection purpose of p2p calling , so here it is decoded and added to p.c for connection
            console.log(
              `payload of type ${p.type} is recieved successfully✅; sent by user${p.sender_id}`,
            );

            // reciever's ice is recieved but not added untill call is not accepeted

            // candidate validation check - check if really incoming custom json.raw payload contains the ice 🧊
            const rawIceCandidate = p.audio_payload_only; //! raw ice ( possible networking path - "34.59x udp" )
            if (
              !rawIceCandidate ||
              (!rawIceCandidate.candidate &&
                rawIceCandidate.sdpMid == null &&
                rawIceCandidate.sdpMLineIndex == null)
            ) {
              console.error(
                "recieved ice candidate is either null or invalid;connection could not be created for p2p calling❌",
              );
              break;
            }

            // if ice is coming with correct information for connection
            // decode ice to get possible connection route -> add ice -> exchange ice -> if passed -> connected✅
            if (
              peerConnection.current &&
              peerConnection.current.remoteDescription
            ) {
              // * since connection will be made and streams would be served -> p.c should be there for accessing streams && remoteDesc should be already set up for acknowledging and decoding incoming stream
              // if both are there and exists
              const webRtcRecognizedIce = new RTCIceCandidate(rawIceCandidate);
              await peerConnection.current.addIceCandidate(webRtcRecognizedIce);
              console.log(
                "successfully added ice candidate to the peer connection✅;p2p calling connection successfully created✅ ",
              );
              break;
            } else {
              //! note - if suppose on reciever side -> ice is gathered and recieved here through ws sent payload of type "ice-candidate"
              //   ! but connect is not there and also the remote description is not set up
              //! delay addding untill call is not accept,

              //   pushing into the queue arr
              queueRecieverCandidates.current.push(p.audio_payload_only);
              console.log(
                "successfully pushed candidate to the queue; call has not accepted yet...",
              );
              break;
            }
          } // ..ice-candidate
          case "hangup": {
            console.log(
              `payload of type ${p.type} is recieved successfully✅; sent by user${p.sender_id}`,
            );

            //  do the peer connections and refs cleanup
            console.log(
              "cleaning up loose connections and stuff;as call is hanged up...",
            );

            handleCleanup(); //* invoking this to cleanup

            setCallState("idle");
            console.log(
              "user has been successfully disconnected from the p2p calling connection and all cleanup is done✅.",
            );
          } // ..hangup
        }
      }); //..call unsubscribe when ws err

      return () => {
        if (unsubscribe) {
          unsubscribe();
        }
      };
    }, [subscribeNotifications, passedCurrentUserID]); //* dependency arr which cause re-render

    // accept incoming call
    const acceptCall = async () => {
      console.log("accepting incoming video call...");

      // this is where we do the all flow again -> accept call to recieve remote streams and sending streams and all

      try {
        // 1. get rtc conn, loaded from rtc config
        setCallState("active");
        const pconn = new RTCPeerConnection(rtcConfig);
        peerConnection.current = pconn;
        if (peerConnection.current) {
          console.log("peer connection is successfully intialized✅;");
        }

        // register all webrtc handlers first - onicecandidate,ontrack
        peerConnection.current.onicecandidate = (event) => {
          if (event.candidate && sendNotifications) {
            // * candidate becomes not nill, means ice is gathered -> send to reciever -> to let other side decode it
            //  and get -> webrtc recognized candidate -> ice exchanges -> once,compeleted -> connected✅
            const candidatesdp = event?.candidate;
            const outboundCandidatePayloadEvent = {
              type: "ice-candidate",
              sender_id: passedCurrentUserID,
              reciever_id: targetPeerID.current,
              content: "video",
              audio_payload_only: candidatesdp,
            };

            sendNotifications(outboundCandidatePayloadEvent);
            console.log(
              "successfully gathered ice~candidate✅;sent to the reciever successfully ✅;remote desc must be set up on other side for ingres.",
            );
          } else {
            console.warn("failed to gather ice candidate");
          }
        };

        peerConnection.current.ontrack = (event) => {
          const remoteVideoStream = event.streams[0]; //grabbing stream from arr first el

          console.log("---- remote incoming streams check ----");
          console.log("logged remote media video stream details :", {
            active_remote_StreamID: remoteVideoStream.id,
            current_remote_StreamKind: remoteVideoStream.kind, // tells if its audio/video , important property
            current_remote_StreamLabel: remoteVideoStream.label, // logs facecm hd etc/name.
            has_remote_StreamEnabled: remoteVideoStream.enabled,
            has_remote_StreamMuted: remoteVideoStream.muted,
            current_remote_StreamState: remoteVideoStream.state,
          });

          // accessing dom nodes for playing remote streams - video will play both audio + video, don't need seperate stream
          let videRefEl = remoteVideoRef?.current;
          // if both streams exists -> safely display video stream and play audio stream
          if (remoteVideoStream) {
            // we have now both streams insiide the e.streams[0]
            // just need to attach to the refs to play them

            videRefEl.srcObject = remoteVideoStream; // sourced stream
          }
        };

        // 2.set remote desc first for setting up incoming stream config codecs recognition by the pconn (aka peerConnection.current) **need on track
        //   ! it needs incoming offer sdp
        await peerConnection?.current.setRemoteDescription(
          // ! must use - new RTCSessionDescription(plainSdp) -> converts into webrtc recongnized sdp
          new RTCSessionDescription(incomingOfferSdp), //! to set up remote desc -> must pass incomingOfferSdp which carries that info about incoming stream
          //* tells what is incoming recieved from "offer" when <- happens before we hit accept
        );

        //  ~ flush queued candidates after remote desc is setup
        if (
          peerConnection?.current &&
          queueRecieverCandidates.current.length > 0 // candidates exists
        ) {
          // add them first
          for (const queuedCand of queueRecieverCandidates.current) {
            //! skipping null
            if (
              !queuedCand ||
              (!queuedCand.candidate &&
                queuedCand.sdpMid == null &&
                queuedCand.sdpMLineIndex == null)
            ) {
              console.log("skipping null candidates🛑.");
              continue;
            }

            try {
              console.log(
                "adding queued ice candidate to the peer connection.",
              );
              await peerConnection.current.addIceCandidate(
                new RTCIceCandidate(queuedCand),
              );
              console.log("successfully added ice candidate (reciever)✅");
            } catch (err) {
              console.error("failed to add queued candidate(reciever)❌");
            }
            // * only adding right one
          }

          queueRecieverCandidates.current = []; // flush all
          console.log("ice candidate queue is now cleared.");
          console.log(
            "cleared candidate queue;reciever ice candidate is added ✅",
          );
        }

        // ~ constrains -> custom constraint for video configurations
        const maxQualityMediaConstraints = {
          audio: true,
          video: {
            width: { ideal: 1920 },
            height: { ideal: 1080 },
            frameRate: { ideal: 30 }
          }
        };

        const standardMediaConstraints = {
          audio: true,
          video: {
            width: {
              min: 480,
              ideal: 720,
              max: 1920,
            },
            height: {
              min: 480,
              ideal: 720,
              max: 1920,
            },
            frameRate: {
              min: 24,
              ideal: 30,
              max: 60,
            },
          },
        };

        //3.grab local media streams
        const mediaStream = await navigator.mediaDevices.getUserMedia(maxQualityMediaConstraints);
        // store in localstream.current
        //   ! bug - can't use optional chaining on left side
        //   fixed - removed '? '
        localstream.current = mediaStream;
        //4. add stream to the peer connection
        localstream?.current.getTracks().forEach((eachTrackStream) => {
          peerConnection.current.addTrack(eachTrackStream, localstream.current);
        });

        // after adding - source local media stream to ref too for own preview
        localVideoRef.current.srcObject = localstream.current;

        //  ans priority order is ==> remote >> local
        //5.   create answer sdp
        //   bug - must await all pconn calls
        //   fixed - awaiting now
        const ansSdp = await peerConnection.current.createAnswer();
        //6. set local desc for letting other side what's incoming and what it expects (negotitations)

        // ! while sending -> make sure you convert it into webRtc recognized sdp -> by using funcs :
        //       new RTCSessionDescription() → for offer/answer SDPs
        // new RTCIceCandidate()       → for ice candidates
        await peerConnection.current.setLocalDescription(
          new RTCSessionDescription(ansSdp), //* sends sdp - which becomes remote desc for reciever
        ); //

        //7. construct outbound ans payload event
        const outboundAnsEvent = {
          sender_id: passedCurrentUserID,
          reciever_id: targetPeerID.current, // need to set it first reffdly
          type: "answer",
          content: "video",
          audio_payload_only: ansSdp,
        };
        // 8. sending event via ws connection
        if (sendNotifications) {
          sendNotifications(outboundAnsEvent);
        } else {
          console.warn("sendNotifications function not found");
        }
      } catch (err) {
        console.error(err);
      }
    };
    // intialization of rtc connection -> onicecandidates + queue + ontrack+ setting up remoteDesc + ans + local desc
    //*when fired up initalize calling
    const initialiseCalling = async (receiverID) => {
      // *findmeFuncLogic

      try {
        targetPeerID.current = receiverID;
        setCallState("calling"); //* for loading ui instantly
        console.log(
          "caller has found targetted userID for intiating video call :",
          targetPeerID.current,
        );
        //**1. rtc conn from cofig
        const pconn = new RTCPeerConnection(rtcConfig);

        //**2. store peer connection in ref✅
        peerConnection.current = pconn;

        //$ registering webRTC handlers once rtc connection is up //

        // onicecandidate => gathers ice and once gathering is done -> sends ice (possible connection route/path ) candidate ( in plain js) to the reciever
        peerConnection.current.onicecandidate = (event) => {
          // # if ice 🧊 gathering is done,{ found where to locate client and its internet ip addr and all}, send ice to the reciever
          if (event.candidate && sendNotifications) {
            // * candidate becomes not nill, means ice is gathered -> send to reciever -> to let other side decode it
            //  and get -> webrtc recognized candidate -> ice exchanges -> once,compeleted -> connected✅
            const candidatesdp = event?.candidate;
            const outboundCandidatePayloadEvent = {
              type: "ice-candidate",
              sender_id: Number(passedCurrentUserID),
              reciever_id: Number(targetPeerID.current),
              content: "video",
              audio_payload_only: candidatesdp,
            };

            sendNotifications(outboundCandidatePayloadEvent);
            console.log(
              "successfully gathered ice~candidate✅;sent to the reciever successfully ✅;remote desc must be set up on other side for ingres.",
            );
          } else {
            console.warn("failed to gather ice candidate");
          }
        };

        // ontrack => access remote incoming stream and play via video ref

        peerConnection.current.ontrack = (event) => {
          // ! remember - ontrack fired automatically exchanges are done -> caller sends offer -> if reciever accepts and add media stream to the pconn <- on track fires
          // ! that's why we need these handlers both sides cause -> when recievers accepts -> it's ontrack fires -> plays remote incoming stream that was added by the caller and set through localdesc
          // ! that's why reciever must set remoteDesc cause -> stream is already in pconn and ready to be played but -> if not remote desc (remote incoming stream known to pc) -> playback won't happen
          // remote streams are stored and returned by e.streams -> returns arr in
          // !note - e.track -> gives the remote track stream (audio) and event.streams (arr) -> has remote video stream

          // bug - cannot use e.streams[0] -> it contains full media stream, not seperate audio
          // fixed - attached fulll stream on it -> sourced it
          const remoteVideoStream = event.streams[0]; //grabbing stream from arr first el

          console.log("---- remote incoming streams check ----");
          console.log("logged remote media video stream details :", {
            active_remote_StreamID: remoteVideoStream.id,
            current_remote_StreamKind: remoteVideoStream.kind, // tells if its audio/video , important property
            current_remote_StreamLabel: remoteVideoStream.label, // logs facecm hd etc/name.
            has_remote_StreamEnabled: remoteVideoStream.enabled,
            has_remote_StreamMuted: remoteVideoStream.muted,
            current_remote_StreamState: remoteVideoStream.state,
          });

          // accessing dom nodes for playing remote streams - video will play both audio + video, don't need seperate stream
          let videRefEl = remoteVideoRef?.current;
          // if both streams exists -> safely display video stream and play audio stream
          if (remoteVideoStream) {
            // we have now both streams insiide the e.streams[0]
            // just need to attach to the refs to play them

            videRefEl.srcObject = remoteVideoStream; // sourced stream ( full media stream -> vid + aud)
          }
        };

        //$ ...handlers //

        //**3. since we are intializing call, we have to handle outgoing webrtc operations first

        // grab media stream
        const mediaStream = await navigator.mediaDevices.getUserMedia({
          audio: true,
          // ! setting true for video -> capturing video too, but instead direct true, we pass constrainted config -> for safely caoturing stream
          video: {
            // setting to use 720p resolution
            width: {
              // using robust min,ideal,max for flawless capture
              min: 480,
              ideal: 720,
              max: 1920,
            },
            height: {
              min: 480,
              ideal: 720,
              max: 1920,
            },
            frameRate: {
              min: 24,
              ideal: 30,
              max: 60,
            },
          },
        }); // ..mediaStream

        if (mediaStream) {
          console.log(
            "captured local media stream📸;storing in localstream...",
          );
          localstream.current = mediaStream;
          console.log("captured media stream details", {
            mediaStreamID: mediaStream.id,
          });
        }
        console.log(
          "successfully grabbed media stream✅; yet to add into the p.c",
        );

        // local streams get added to getTracks arr in index postioning of el -> [0]
        // im saying streams cause  both audio and video are opened, so get tracks return arr of both streams where
        // audio takes the first posi, and video being second but in terms of arr el's index [0-audio,1-video]

        const localVideoStreamCheck = localstream.current?.getTracks()[1]; //? for not running into runtime undefined  errors
        console.log("---- starting local stream check of both streams-----");
        console.log("logged client's video media stream details :", {
          activeStreamID: localVideoStreamCheck?.id,
          currentStreamKind: localVideoStreamCheck?.kind, // tells if its audio/video , important property
          currentStreamLabel: localVideoStreamCheck?.label, // logs facecm hd etc/name.
          hasStreamEnabled: localVideoStreamCheck?.enabled,
          hasStreamMuted: localVideoStreamCheck?.muted,
          currentStreamState: localVideoStreamCheck?.state,
        });

        const localAudioStreamCheck = localstream.current?.getTracks()[0]; //? for not running into runtime undefined  errors
        console.log("logged client's audio track stream details :", {
          activeStreamID: localAudioStreamCheck?.id,
          currentStreamKind: localAudioStreamCheck?.kind, // tells if its audio/video , important property
          currentStreamLabel: localAudioStreamCheck?.label, // logs facecm hd etc/name.
          hasStreamEnabled: localAudioStreamCheck?.enabled,
          hasStreamMuted: localAudioStreamCheck?.muted,
          currentStreamState: localAudioStreamCheck?.state,
        });

        // once both streams are verified
        // **4. we store both in peer connection (p.c -> source like hub where things get stored and acknowledged for exchange)
        console.log(
          "local streams verified✅;adding streams to the peer connection...",
        );
        localstream.current.getTracks().forEach((eachLocalStreamTrack) => {
          // storing stream in p.c for remote play on the other side
          peerConnection.current.addTrack(
            eachLocalStreamTrack,
            localstream.current,
          );
          console.log(
            `successfully added ${eachLocalStreamTrack.kind} stream to the peer connection`,
          );
        });

        // add local stream to the local video ref
        localVideoRef.current.srcObject = localstream.current; //* attached local media stream to display it

        //** 5 - create offer
        const offerSdp = await peerConnection.current.createOffer(); // must have opened localstreeams so offer carries that information and tells recievers what it expects

        //** 6. setting up local description to tell what's coming on other side and what offer it expects
        await peerConnection.current.setLocalDescription(offerSdp);

        const outboundOfferEvent = {
          sender_id: passedCurrentUserID,
          reciever_id: targetPeerID.current,
          type: "offer",
          content: "video",
          audio_payload_only: offerSdp,
        };
        //** 7. send offer attached payload via ws connection
        if (sendNotifications) {
          sendNotifications(outboundOfferEvent);
          console.log(
            "video call initialised with user ID:",
            outboundOfferEvent.sender_id,
          );
        } else {
          console.warn("sendNotifications function not found");
        }
      } catch (err) {
        console.error(err);
        handleCleanup(); //* do cleanup if something went wrong to intialise call but things would have got loaded
      }
    };

    // Expose initialiseCalling trigger to parent component
    useImperativeHandle(ref, () => ({
      initialiseCalling: (receiverID) => {
        initialiseCalling(receiverID);
      },
    }));

    // Effect to toggle local audio tracks (mute/unmute)
    useEffect(() => {
      if (localstream.current) {
        const audioTracks = localstream.current.getAudioTracks();
        audioTracks.forEach((track) => {
          track.enabled = !isMuted;
          console.log(
            `Local audio track ${track.id} enabled state set to: ${!isMuted}`,
          );
        });
      }
    }, [isMuted]);

    // Effect to toggle local video tracks (camera on/off)
    useEffect(() => {
      if (localstream.current) {
        const videoTracks = localstream.current.getVideoTracks();
        videoTracks.forEach((track) => {
          track.enabled = !isVideoOff;
          console.log(
            `Local video track ${track.id} enabled state set to: ${!isVideoOff}`,
          );
        });
      }
    }, [isVideoOff]);

    const peerLabel = targetPeerID.current || "User";

    return (
      <>
        {/* always mounting playback refs video players -> remote + local */}

        {/* ..end.. */}

        {/* always mounting overlay to preserve non-null refs, showing/hiding via display property */}
        <div
          className="videocall-overlay"
          style={{ display: callState === "idle" ? "none" : "flex" }}
        >
          <div className="videocall-container">
            {/* 1. Large Remote Stream Container */}
            <div className="videocall-remote-container">
              <video
                ref={remoteVideoRef}
                className="videocall-remote-video"
                autoPlay
                playsInline
                style={{ display: callState === "active" ? "block" : "none" }}
              />
              {callState !== "active" && (
                <div className="videocall-remote-placeholder">
                  <div className="videocall-remote-avatar">
                    {String(peerLabel).substring(0, 2).toUpperCase()}
                  </div>
                  <h3 className="videocall-remote-name">User #{peerLabel}</h3>
                  <p className="videocall-remote-status">
                    {callState === "incoming"
                      ? "Incoming Call..."
                      : "Calling..."}
                  </p>
                </div>
              )}
            </div>

            {/* 2. Floating Medium Own Stream (PIP) Container */}
            <div className="videocall-own-container">
              <video
                ref={localVideoRef}
                className="videocall-own-video"
                autoPlay
                playsInline
                muted
                style={{ display: isVideoOff ? "none" : "block" }}
              />
              {isVideoOff && (
                <div className="videocall-own-placeholder">
                  <div className="videocall-own-avatar">
                    {String(passedCurrentUserID || "ME")
                      .substring(0, 2)
                      .toUpperCase()}
                  </div>
                  <span className="videocall-own-label">You</span>
                </div>
              )}
              {isMuted && (
                <span className="videocall-muted-indicator" title="Muted">
                  🎤❌
                </span>
              )}
            </div>

            {/* 3. Call Details Header */}
            <div className="videocall-header">
              <div className="videocall-info">
                <span className="videocall-badge">LIVE VIDEO CALL</span>
                <h2>Active Connection: User #{peerLabel}</h2>
              </div>
            </div>

            {/* 4. Sleek Call Controls Bar */}
            <div className="videocall-controls-bar">
              {callState === "incoming" ? (
                <>
                  {/* Accept Call Button */}
                  <button
                    onClick={acceptCall}
                    className="videocall-control-btn btn-accept"
                    title="Accept Video Call"
                  >
                    📹
                  </button>
                  {/* Decline Call Button */}
                  <button
                    onClick={handleHangup}
                    className="videocall-control-btn btn-hangup"
                    title="Decline Video Call"
                  >
                    ✖
                  </button>
                </>
              ) : callState === "calling" ? (
                /* Cancel Call Button */
                <button
                  onClick={handleHangup}
                  className="videocall-control-btn btn-hangup"
                  title="Cancel Call"
                >
                  🛑
                </button>
              ) : (
                <>
                  {/* Mic Toggle */}
                  <button
                    onClick={() => setIsMuted(!isMuted)}
                    className={`videocall-control-btn ${isMuted ? "is-muted" : ""}`}
                    title={isMuted ? "Unmute Mic" : "Mute Mic"}
                  >
                    {isMuted ? "🎙️❌" : "🎙️"}
                  </button>

                  {/* End Call / Hangup */}
                  <button
                    onClick={handleHangup}
                    className="videocall-control-btn btn-hangup"
                    title="End Video Call"
                  >
                    🛑
                  </button>

                  {/* Video Toggle */}
                  <button
                    onClick={() => setIsVideoOff(!isVideoOff)}
                    className={`videocall-control-btn ${isVideoOff ? "is-video-off" : ""}`}
                    title={isVideoOff ? "Turn Camera On" : "Turn Camera Off"}
                  >
                    {isVideoOff ? "📹❌" : "📹"}
                  </button>
                </>
              )}
            </div>
          </div>
        </div>
      </>
    );
  },
);

VideoCalling.displayName = "VideoCalling";
export default VideoCalling;
