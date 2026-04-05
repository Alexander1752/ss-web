import React from 'react';
import { View, Text, StyleSheet } from 'react-native';

const RegisterScreen: React.FC = () => {
  return (
    <View style={styles.container}>
      <Text style={styles.title}>Register</Text>
      <Text style={styles.description}>
        Registration flow will be migrated next. This screen is a placeholder for now.
      </Text>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    padding: 24,
    justifyContent: 'center',
    backgroundColor: '#f6f8fb'
  },
  title: {
    fontSize: 24,
    fontWeight: '700',
    color: '#1d2433',
    marginBottom: 8
  },
  description: {
    color: '#5d6a85',
    fontSize: 16,
    lineHeight: 24
  }
});

export default RegisterScreen;
