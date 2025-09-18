import React, { useState } from 'react';
import { Link, useLocation, Outlet } from 'react-router-dom';
import { Layout, Menu, Typography } from 'antd';
import {
    DatabaseOutlined,
    FileTextOutlined,
    ScheduleOutlined,
} from '@ant-design/icons';

const { Header, Content, Footer, Sider } = Layout;
const { Title } = Typography;

const menuItems = [
    {
        key: '/datasources',
        icon: <DatabaseOutlined />,
        label: <Link to="/datasources">資料來源管理</Link>,
    },
    {
        key: '/reports',
        icon: <FileTextOutlined />,
        label: <Link to="/reports">報表定義</Link>,
    },
    {
        key: '/schedules',
        icon: <ScheduleOutlined />,
        label: <Link to="/schedules">排程管理</Link>,
    },
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
                <Menu theme="dark" selectedKeys={[getSelectedKey()]} mode="inline" items={menuItems} />
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
