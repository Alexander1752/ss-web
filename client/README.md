# Security of Systems - First Force / LF4A - Delta Force

A front-end application for a university project that serves as a security system interface. The application captures photos from connected devices, extracts text using OCR, and enables searching based on the extracted text content.

## Project Structure

```
client/
├── public/
├── src/
│   ├── assets/           # Static assets like images and fallbacks
│   ├── components/       # Reusable UI components
│   │   ├── devicesCards/ # Cards for device display
│   │   ├── photosCards/  # Cards for photo display
│   │   └── ...          # Other components
│   ├── contexts/         # React contexts for global state
│   │   └── AuthContext.tsx # Authentication context
│   ├── pages/            # Application pages
│   │   ├── devicesPage/  # Devices management page
│   │   ├── homePage/     # Home page
│   │   ├── loginPage/    # Login page
│   │   ├── photosPage/   # Photos display and search
│   │   └── ...          # Other pages
│   ├── App.tsx           # Main application component
│   ├── index.tsx         # Entry point
│   └── Router.tsx        # Application routing
├── package.json
└── tailwind.config.js    # Tailwind CSS configuration
```

## Key Components

- **Authentication System**: JWT-based authentication with token storage in localStorage
- **Photo Card Component**: Displays images with extracted text and zoom functionality
- **Device Card Component**: Displays device information with mode switching controls
- **Navigation Bar**: App header with conditional navigation based on authentication state
- **Protected Routes**: Routes that require authentication to access

## Technologies Used

- **React**: Front-end UI library
- **TypeScript**: Type-safe JavaScript
- **React Router**: For application routing
- **TailwindCSS**: For styling and responsive design
- **JWT**: For authentication
- **LocalStorage**: For persisting user data and search preferences

## Getting Started

### Prerequisites

- Node.js (v16 or later) (for development v24.0.2 was used)
- npm or yarn (for development npm - v11.3.0 and yarn - v1.22.22 were used)

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/your-username/security-system.git
   cd security-system/client
   ```

2. Install dependencies:
   ```bash
   yarn install
   ```

3. (Optional) Create a `.env` file (or `.env.local`) to point the client at a custom backend:
   ```bash
   VITE_API_BASE_URL=https://your-api-hostname
   ```
   If this variable is not provided during local development the app will use `http://localhost:8080` automatically.

### Running

To start the development server:

```bash
yarn dev:poll
```

The application will be available at `http://localhost:3000`.

### Building

To create a production build:

```bash
yarn build
```

The build files will be available in the `build` directory.

## Usage

1. **Login/Register**: Start by logging in with your credentials or registering a new account.
2. **View Photos**: Navigate to the Photos page to see captured images.
3. **Search Photos**: Use the search bar to find photos by extracted text, date range, or device.
4. **Manage Devices**: Access the Devices page to see all connected devices.
5. **Control Devices**: Toggle devices between Normal and Live modes as needed.

## Features

- **User Authentication**: Secure login and registration
- **Photo Browse**: View photos captured by connected devices
- **Advanced Search**: Search photos by text, date range, and device
- **Device Management**: View and control all connected security devices
- **Mode Switching**: Switch devices between Normal and Live modes
- **Persistent Settings**: Search preferences are preserved across sessions
- **Responsive Design**: Works on mobile, tablet, and desktop devices

## Troubleshooting

### Authentication Issues

- **Problem**: Unable to login
  **Solution**: Verify your credentials and check your internet connection. If the issue persists, try clearing your browser cache.

- **Problem**: Getting logged out unexpectedly
  **Solution**: Your session might have expired. Login again to get a new token.

### Photo Loading Issues

- **Problem**: Photos not displaying
  **Solution**: The application uses fallback images when the source URL is invalid. Check your internet connection or try refreshing the page.

- **Problem**: Search returns no results
  **Solution**: Try broadening your search criteria or check if you have the correct permissions to view those photos.

### Device Control Issues

- **Problem**: Unable to switch device modes
  **Solution**: Verify you have the necessary permissions. Try refreshing the page and checking your internet connection.

- **Problem**: Devices not showing up
  **Solution**: Ensure you're logged in with an account that has access to device management.
