import React from 'react';
import MainLayout from './components/layouts/MainLayout';
import ReportDefinitionPage from './pages/ReportDefinition';
import './App.css';

const App: React.FC = () => {
    return (
        <MainLayout>
            <ReportDefinitionPage />
        </MainLayout>
    );
};

export default App;
