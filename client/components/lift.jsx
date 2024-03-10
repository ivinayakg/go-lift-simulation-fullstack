import { useEffect, useState } from "react";
import { View, StyleSheet, Animated } from "react-native";

export default function Lift({ liftData, changeFloorSetter }) {
  const [moveLift] = useState(new Animated.Value(25));
  const [closeDoor] = useState(new Animated.Value(32));

  function changeFloor(request, removeEvent) {
    let floorDiff = Math.abs(request.floorToReach - request.currentFloor);
    Animated.timing(moveLift, {
      toValue: (request.floorToReach - 1) * -125,
      duration: 2000 * floorDiff,
      useNativeDriver: true,
    }).start();

    //door animation
    let doorAnimation = Animated.sequence([
      Animated.timing(closeDoor, {
        toValue: 0,
        duration: 2500,
        useNativeDriver: false,
        delay: 2000 * floorDiff,
      }),
      Animated.timing(closeDoor, {
        toValue: 32,
        duration: 2500,
        useNativeDriver: false,
        delay: 2000,
      }),
    ]);

    doorAnimation.start();
    if (removeEvent) setTimeout(removeEvent, 2000 * floorDiff + 5000);
  }

  useEffect(() => {
    changeFloorSetter(changeFloor, liftData._id);
    if (liftData.currentFloor === 0) {
      Animated.timing(moveLift, {
        toValue: (liftData.currentFloor) * -125,
        duration: 0,
        useNativeDriver: true,
      }).start();
    } else {
      Animated.timing(moveLift, {
        toValue: (liftData.currentFloor - 1) * -125,
        duration: 0,
        useNativeDriver: true,
      }).start();
    }
  }, []);

  return (
    <Animated.View
      style={[
        styles.lift,
        {
          transform: [{ translateY: moveLift }],
        },
      ]}
    >
      <Animated.View
        style={[styles.leftDoor, { width: closeDoor }]}
      ></Animated.View>
      <Animated.View
        style={[styles.rightDoor, { width: closeDoor }]}
      ></Animated.View>
    </Animated.View>
  );
}

const styles = StyleSheet.create({
  lift: {
    height: 80,
    width: 66,
    backgroundColor: "#000",
    display: "flex",
    justifyContent: "space-between",
    alignItems: "center",
    flexDirection: "row",
  },
  leftDoor: {
    height: 80,
    width: 32,
    backgroundColor: "#555",
  },
  rightDoor: {
    height: 80,
    width: 32,
    backgroundColor: "#555",
  },
});
