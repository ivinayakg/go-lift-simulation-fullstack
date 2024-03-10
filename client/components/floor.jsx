import { Button } from "react-native";
import { View, StyleSheet } from "react-native";

export default function Floor({ first, last, index, lifts, jumpToFloor }) {
  const mainStyles = [styles.floor];
  if (first) mainStyles.push(styles.floor_first);
  if (last) mainStyles.push(styles.floor_last);

  function requestLift() {
    jumpToFloor(index + 1);
  }

  return (
    <View style={mainStyles}>
      <View style={styles.floor_buttonWrapper}>
        <Button
          styles={styles.floor_upperButton}
          title="->"
          onPress={requestLift}
        />
        {/* <Button
          styles={styles.floor_upperButton}
          title="^"
          onPress={requestLift}
        /> */}
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  floor: {
    height: 125,
    borderBottomWidth: 2,
    borderBottomColor: "#333",
    position: "relative",
    display: "flex",
    alignItems: "center",
    flex: 1,
  },
  floor_first: {
    borderTopColor: "#333",
    borderTopWidth: 1,
  },
  floor_last: {
  },
  floor_lowerButton: {
    paddingVertical: 5,
    paddingHorizontal: 9,
    transform: [{ rotateZ: "180deg" }],
  },
  floor_upperButton: {
    paddingVertical: 5,
    paddingHorizontal: 9,
  },
  floor_buttonWrapper: {
    flexDirection: "row",
    justifyContent: "center",
    alignItems: "center",
    borderWidth: 0,
    backgroundColor: "transparent",
    marginHorizontal: 8,
    position: "absolute",
    top: "40%",
    left: 10,
    gap: 10,
  },
});
