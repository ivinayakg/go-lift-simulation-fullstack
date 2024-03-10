import { useState } from "react";
import { Button, TextInput, View, StyleSheet, Text } from "react-native";

const initialInputText = {
  liftInput: 0,
  floorInput: 0,
  sessionIdInput: null,
};

export default function HeaderInput({ liftState, updateState }) {
  const [inputValue, setInputValue] = useState(initialInputText);

  function onSubmit() {
    updateState(inputValue);
  }

  return (
    <View style={styles.header}>
      <View style={styles.header_input}>
        <TextInput
          style={styles.header_input__children}
          onChangeText={(text) =>
            setInputValue((prev) => ({ ...prev, liftInput: text }))
          }
          placeholder="Input Lifts"
          keyboardType="number-pad"
          value={inputValue.liftInput ? String(inputValue.liftInput) : ""}
        />
        <TextInput
          style={styles.header_input__children}
          onChangeText={(text) =>
            setInputValue((prev) => ({ ...prev, floorInput: text }))
          }
          onSubmitEditing={onSubmit}
          placeholder="Input Floors"
          keyboardType="number-pad"
          value={inputValue.floorInput ? String(inputValue.floorInput) : ""}
        />
        <TextInput
          style={styles.header_input__children}
          onChangeText={(text) =>
            setInputValue((prev) => ({ ...prev, sessionIdInput: text }))
          }
          onSubmitEditing={onSubmit}
          placeholder="Input Session Id"
          keyboardType="default"
          value={inputValue.sessionIdInput ?? ""}
        />
        <Button title="Submit" onPress={onSubmit} />
      </View>
      <View style={styles.header_status}>
        <Text>Lifts :- {liftState.lifts.length}</Text>
        <Text>Floors :- {liftState.floors}</Text>
        <Text>Session ID :- {liftState._id}</Text>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  header: {
    width: "100%",
    display: "flex",
    justifyContent: "center",
    alignItems: "center",
    gap: 10,
    flexDirection: "column",
  },
  header_status: {
    flexDirection: "row",
    gap: 10,
  },
  header_input: {
    display: "flex",
    justifyContent: "center",
    alignItems: "center",
    gap: 10,
    flexWrap: "wrap",
  },
  header_input__children: {
    height: 35,
    borderWidth: 1,
    borderColor: "#333",
    width: 200,
    padding: 6,
  },
});
