import React, { useState } from 'react';
import { Link, useLocation, Outlet } from 'react-router-dom';
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
    { key: '/datasources', icon: <DatabaseOutlined />, label: '資料來源管理', path: '/datasources' },
    { key: '/reports', icon: <FileTextOutlined />, label: '報表定義', path: '/reports' },
    { key: '/schedules', icon: <ScheduleOutlined />, label: '排程管理', path: '/schedules' },
    { key: '/history', icon: <HistoryOutlined />, label: '歷史紀錄', path: '/history' },
];

const pageTitles: { [key: string]: string } = {
    '/datasources': '資料來源管理',
    '/reports': '報表定義',
    '/schedules': '排程管理',
    '/history': '歷史紀錄',
};

const MainLayout: React.FC = () => {
    const [collapsed, setCollapsed] = useState(false);
    const location = useLocation();

    const getSelectedKey = () => {
        return location.pathname;
    };

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
                <Menu theme="dark" selectedKeys={[getSelectedKey()]} mode="inline">
                    {menuItems.map(item => (
                        <Menu.Item key={item.key} icon={item.icon}>
                            <Link to={item.path}>{item.label}</Link>
                        </Menu.Item>
                    ))}
                </Menu>
            </Sider>
            <Layout className="site-layout">
                <Header className="site-layout-background" style={{ padding: '0 24px', background: '#fff' }}>
                   <Title level={3} style={{ margin: '16px 0' }}>
                        {pageTitles[location.pathname] || '歡迎'}
                   </Title>
                </Header>
                <Content style={{ margin: '24px 16px 0' }}>
                    <div className="site-layout-background" style={{ padding: 24, minHeight: 360, background: '#fff' }}>
                        <Outlet />
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
