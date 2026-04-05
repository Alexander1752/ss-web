import React from 'react';
import { View, Text, StyleSheet } from 'react-native';

const PhotosScreen: React.FC = () => {
  return (
    <View style={styles.container}>
      <Text style={styles.title}>Photos</Text>
      <Text style={styles.description}>Photos list and upload UI migration is pending.</Text>
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

export default PhotosScreen;
