import React from 'react';
import { View, Text, StyleSheet } from 'react-native';

const StatisticsScreen: React.FC = () => {
  return (
    <View style={styles.container}>
      <Text style={styles.title}>Statistics</Text>
      <Text style={styles.description}>Charts and analytics migration is pending.</Text>
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
    fontSize: 16
  }
});

export default StatisticsScreen;
