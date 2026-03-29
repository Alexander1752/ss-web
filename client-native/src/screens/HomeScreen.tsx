import React from 'react';
import { View, Text, StyleSheet, Pressable } from 'react-native';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import type { RootStackParamList } from '../types/navigation';
import { useAuth } from '../contexts/AuthContext';

type Props = NativeStackScreenProps<RootStackParamList, 'Home'>;

const HomeScreen: React.FC<Props> = ({ navigation }) => {
  const { isLoggedIn, logout } = useAuth();

  return (
    <View style={styles.container}>
      <Text style={styles.title}>SS APP</Text>
      <Text style={styles.subtitle}>First mobile refactor baseline</Text>

      {!isLoggedIn ? (
        <View style={styles.actions}>
          <Pressable style={styles.primaryButton} onPress={() => navigation.navigate('Login')}>
            <Text style={styles.buttonText}>Login</Text>
          </Pressable>
          <Pressable style={styles.secondaryButton} onPress={() => navigation.navigate('Register')}>
            <Text style={styles.secondaryText}>Register</Text>
          </Pressable>
        </View>
      ) : (
        <View style={styles.actions}>
          <Pressable style={styles.primaryButton} onPress={() => navigation.navigate('Photos')}>
            <Text style={styles.buttonText}>Photos</Text>
          </Pressable>
          <Pressable style={styles.primaryButton} onPress={() => navigation.navigate('Devices')}>
            <Text style={styles.buttonText}>Devices</Text>
          </Pressable>
          <Pressable style={styles.primaryButton} onPress={() => navigation.navigate('Statistics')}>
            <Text style={styles.buttonText}>Statistics</Text>
          </Pressable>
          <Pressable
            style={styles.secondaryButton}
            onPress={() => {
              logout();
            }}
          >
            <Text style={styles.secondaryText}>Logout</Text>
          </Pressable>
        </View>
      )}
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f6f8fb',
    padding: 24,
    justifyContent: 'center'
  },
  title: {
    fontSize: 28,
    fontWeight: '700',
    marginBottom: 8,
    color: '#1d2433'
  },
  subtitle: {
    fontSize: 16,
    color: '#5d6a85',
    marginBottom: 28
  },
  actions: {
    gap: 12
  },
  primaryButton: {
    backgroundColor: '#1f6feb',
    borderRadius: 12,
    paddingVertical: 14,
    alignItems: 'center'
  },
  secondaryButton: {
    borderColor: '#1f6feb',
    borderWidth: 1,
    borderRadius: 12,
    paddingVertical: 14,
    alignItems: 'center'
  },
  buttonText: {
    color: '#ffffff',
    fontWeight: '600',
    fontSize: 16
  },
  secondaryText: {
    color: '#1f6feb',
    fontWeight: '600',
    fontSize: 16
  }
});

export default HomeScreen;
