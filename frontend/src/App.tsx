import React from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import MainLayout from './components/layouts/MainLayout';
import DataSourceManagementPage from './pages/DataSourceManagementPage';
import ReportDefinitionPage from './pages/ReportDefinition';
import ScheduleManagementPage from './pages/ScheduleManagementPage';
import HistoryPage from './pages/HistoryPage';

// Import Ant Design CSS
import 'antd/dist/reset.css';

const App: React.FC = () => {
  return (
    <Routes>
      <Route path="/" element={<MainLayout />}>
        <Route index element={<Navigate to="/datasources" replace />} />
        <Route path="datasources" element={<DataSourceManagementPage />} />
        <Route path="reports" element={<ReportDefinitionPage />} />
        <Route path="schedules" element={<ScheduleManagementPage />} />
        <Route path="history" element={<HistoryPage />} />
        {/* Redirect any unknown paths to the main page */}
        <Route path="*" element={<Navigate to="/datasources" replace />} />
      </Route>
    </Routes>
  );
};

export default App;
