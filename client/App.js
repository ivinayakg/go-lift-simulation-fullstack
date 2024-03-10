import { useCallback, useEffect, useState } from "react";
import { StyleSheet, View, ScrollView } from "react-native";
import HeaderInput from "./components/header";
import Lift from "./components/lift";
import Floor from "./components/floor";
import {
  baseurl,
  createRequest,
  createSession,
  fetchSession,
} from "./utils/server";

const initialLiftState = {
  floors: 0,
  lifts: [],
  _id: null,
  clientId: null,
};

export default function App() {
  const socketUrl = "wss://lift-api.ivinayakg.me"
  const [liftState, setLiftState] = useState(initialLiftState);
  const [clientState, setClientState] = useState({ clientId: null });
  const [liftsSetterState, setLiftsSetterState] = useState({});

  const updateLiftSetterFunc = (func, liftId) => {
    setLiftsSetterState((prev) => {
      return { ...prev, [liftId]: func };
    });
  };

  const updateState = async (inputValue) => {
    const { liftInput, floorInput, sessionIdInput } = inputValue;
    if (sessionIdInput) {
      const data = await fetchSession(sessionIdInput);
      setLiftState(data);
    } else {
      const data = await createSession(floorInput, liftInput);
      setLiftState(data);
    }
  };

  const jumpToFloorClicked = async (floorToReach) => {
    let requestData = await createRequest(
      liftState._id,
      clientState.clientId,
      floorToReach
    );
    let lift = requestData.lift;
    if (liftsSetterState[lift._id]) {
      let setFun = liftsSetterState[lift._id];
      setFun({
        currentFloor: lift.currentFloor,
        floorToReach: requestData.requestedFloor,
      });
    } else {
      console.error("lift setter not working properly");
    }
  };

  const jumpToFloorClickedSocket = useCallback(
    (data) => {
      let requestData = data.body;
      let liftId = requestData.lift_id;
      let lift = liftState.lifts.find((lift) => lift._id === liftId);
      if (liftsSetterState[liftId]) {
        let setFun = liftsSetterState[liftId];
        setFun({
          currentFloor: lift.currentFloor,
          floorToReach: requestData.floor_requested,
        });
      } else {
        console.error("lift setter not working properly");
      }
    },
    [liftState, liftsSetterState]
  );

  useEffect(() => {
    let socket;
    let sessionID = liftState._id;
    if (sessionID && sessionID !== "") {
      socket = new WebSocket(
        `${socketUrl}/ws/?sessionId=${sessionID}`
      );
      socket.onmessage = async (event) => {
        const data = JSON.parse(event.data);
        if (data.body.event === "client_info") {
          setClientState((prev) => ({ ...prev, clientId: data.body.clientId }));
        }
        if (data.body.event === "lift_moved") {
          jumpToFloorClickedSocket(data);
        }
      };
    }

    return () => {
      socket?.close();
    };
  }, [liftState, jumpToFloorClickedSocket]);

  return (
    <ScrollView
      style={styles.container}
      contentContainerStyle={[
        { alignItems: "center", justifyContent: "center" },
      ]}
    >
      <HeaderInput liftState={liftState} updateState={updateState} />

      <ScrollView style={styles.main} horizontal={true}>
        <View
          style={{
            width: liftState.lifts.length * 90 + 50,
            height: liftState.lifts.length > 0 ? liftState.floors * 125 : 0,
            flexDirection: "column",
          }}
        >
          <View style={styles.lift_wrapper}>
            {liftState.lifts.map((lift) => {
              return (
                <Lift
                  liftData={lift}
                  key={lift._id}
                  changeFloorSetter={updateLiftSetterFunc}
                />
              );
            })}
          </View>
          {[...Array(Number(liftState.floors))].map((floor, index) => (
            <Floor
              first={index === 0}
              last={index === Number(liftState.floors) - 1}
              index={Number(liftState.floors) - 1 - index}
              key={index}
              lifts={liftState.lifts.length}
              jumpToFloor={jumpToFloorClicked}
            />
          ))}
        </View>
      </ScrollView>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {
    display: "flex",
    backgroundColor: "#fff",
    paddingTop: "10%",
    marginTop: "10%",
  },
  lift_wrapper: {
    position: "absolute",
    width: "100%",
    height: "100%",
    padding: 10,
    display: "flex",
    justifyContent: "flex-end",
    alignItems: "flex-end",
    gap: 20,
    flexDirection: "row",
  },
  main: {
    margin: 8,
    fix: 1,
  },
});
