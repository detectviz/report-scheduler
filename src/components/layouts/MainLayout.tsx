import React, { useState } from 'react';
import { Layout, Menu, Typography } from 'antd';
import {
    DatabaseOutlined,
    FileTextOutlined,
    ScheduleOutlined,
    HistoryOutlined,
} from '@ant-design/icons';

const { Header, Content, Footer, Sider } = Layout;
const { Title } = Typography;

const menuItems = [
    { key: 'datasources', icon: <DatabaseOutlined />, label: '資料來源管理' },
    { key: 'reports', icon: <FileTextOutlined />, label: '報表定義' },
    { key: 'schedules', icon: <ScheduleOutlined />, label: '排程管理' },
    { key: 'history', icon: <HistoryOutlined />, label: '歷史紀錄與稽核' },
];

const pageTitles: { [key: string]: string } = {
    'datasources': '資料來源管理',
    'reports': '報表定義',
    'schedules': '排程管理',
    'history': '歷史紀錄與稽核',
};

interface MainLayoutProps {
    children: React.ReactNode;
}

const MainLayout: React.FC<MainLayoutProps> = ({ children }) => {
    const [collapsed, setCollapsed] = useState(false);
    const [selectedKey, setSelectedKey] = useState('reports'); // Default to reports page

    return (
        <Layout style={{ minHeight: '100vh' }}>
            <Sider collapsible collapsed={collapsed} onCollapse={setCollapsed}>
                <div style={{
                    height: '32px',
                    margin: '16px',
                    background: 'rgba(255, 255, 255, 0.3)',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    color: 'white',
                    fontSize: '16px',
                    fontWeight: 'bold',
                }}>
                    報表系統
                </div>
                <Menu
                    theme="dark"
                    selectedKeys={[selectedKey]}
                    mode="inline"
                    items={menuItems}
                    onClick={({ key }) => setSelectedKey(key)}
                />
            </Sider>
            <Layout>
                <Header style={{ padding: '0 24px', background: '#fff' }}>
                    <Title level={3} style={{ margin: '16px 0' }}>
                        {pageTitles[selectedKey]}
                    </Title>
                </Header>
                <Content style={{ margin: '24px 16px 0' }}>
                    <div style={{ padding: 24, minHeight: 360, background: '#fff' }}>
                        {children}
                    </div>
                </Content>
                <Footer style={{ textAlign: 'center' }}>
                    Report Scheduler ©{new Date().getFullYear()} Created with Ant Design
                </Footer>
            </Layout>
        </Layout>
    );
};

export default MainLayout;
