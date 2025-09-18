import React, { useState, useEffect, useRef, useCallback } from 'react';
import { Form, Input, Select, Divider, Transfer, Button, message, Typography, Table, Spin, Space } from 'antd';
import { DndProvider, useDrag, useDrop } from 'react-dnd';
import { HTML5Backend } from 'react-dnd-html5-backend';
import { useNavigate, useParams } from 'react-router-dom';
import update from 'immutability-helper';

import { getReportDefinitionById, createReportDefinition, updateReportDefinition } from '../api/report';
import type { ReportElement } from '../api/report';
import { getDataSources, getDataSourceElements } from '../api/dataSource';
import type { DataSource, AvailableElement } from '../api/dataSource';

const { Title } = Typography;
const { Option } = Select;

// Draggable Item Component
const type = 'DraggableItem';
interface DraggableItemProps {
    index: number;
    id: string;
    text: string;
    moveItem: (dragIndex: number, hoverIndex: number) => void;
}

const DraggableItem: React.FC<DraggableItemProps> = ({ id, text, index, moveItem }) => {
    const ref = useRef<HTMLDivElement>(null);
    const [, drop] = useDrop({
        accept: type,
        hover(item: { index: number }) {
            if (!ref.current) return;
            const dragIndex = item.index;
            const hoverIndex = index;
            if (dragIndex === hoverIndex) return;
            moveItem(dragIndex, hoverIndex);
            item.index = hoverIndex;
        },
    });
    const [{ isDragging }, drag] = useDrag({
        type,
        item: { id, index },
        collect: (monitor) => ({ isDragging: monitor.isDragging() }),
    });
    drag(drop(ref));
    return (
        <div ref={ref} style={{ padding: '8px 12px', marginBottom: 4, backgroundColor: 'white', border: '1px solid #d9d9d9', borderRadius: 4, cursor: 'move', opacity: isDragging ? 0.5 : 1 }}>
            {text}
        </div>
    );
};


const ReportDefinitionForm: React.FC = () => {
    const [form] = Form.useForm();
    const navigate = useNavigate();
    const { id } = useParams<{ id: string }>();

    const [dataSources, setDataSources] = useState<DataSource[]>([]);
    const [availableElements, setAvailableElements] = useState<AvailableElement[]>([]);
    const [elementsLoading, setElementsLoading] = useState(false);
    const [targetKeys, setTargetKeys] = useState<string[]>([]);

    const [loading, setLoading] = useState(false);
    const [isSaving, setIsSaving] = useState(false);

    const selectedDataSourceId = Form.useWatch('datasource_id', form);

    // Fetch Data Sources
    useEffect(() => {
        const fetchDS = async () => {
            try {
                const ds = await getDataSources();
                setDataSources(ds || []);
            } catch (error) {
                message.error('無法獲取資料來源列表');
            }
        };
        fetchDS();
    }, []);

    // Fetch Report Definition if in edit mode
    useEffect(() => {
        if (id) {
            setLoading(true);
            getReportDefinitionById(id).then(data => {
                const { time_range, ...restData } = data;
                form.setFieldsValue(restData);

                // Parse time_range to populate the UI
                const quickRanges = ["now-1h", "now-24h", "now-7d", "now-30d"];
                if (quickRanges.includes(time_range)) {
                    form.setFieldsValue({ time_range_quick: time_range, time_range: time_range });
                } else if (time_range) {
                    const match = time_range.match(/now-(\d+)([mhdwyM])/);
                    if (match) {
                        form.setFieldsValue({
                            time_range_custom_value: parseInt(match[1], 10),
                            time_range_custom_unit: match[2],
                            time_range: time_range,
                        });
                    }
                }

                setTargetKeys(data.elements.map(el => el.id));
                setLoading(false);
            }).catch(error => {
                console.error("Failed to fetch report definition", error);
                setLoading(false);
            });
        }
    }, [id, form]);

    // Fetch available elements when data source changes
    useEffect(() => {
        if (selectedDataSourceId) {
            setElementsLoading(true);
            getDataSourceElements(selectedDataSourceId).then(elements => {
                setAvailableElements(elements || []);
            }).catch(() => {
                message.error('無法獲取可選項目列表');
                setAvailableElements([]);
            }).finally(() => {
                setElementsLoading(false);
            });
        } else {
            setAvailableElements([]);
        }
    }, [selectedDataSourceId]);

    const onSave = async () => {
        try {
            const values = await form.validateFields();
            setIsSaving(true);

            const elements: ReportElement[] = targetKeys.map((key, index) => {
                const item = availableElements.find(el => el.id === key);
                return {
                    id: key,
                    type: item?.type || 'dashboard',
                    title: item?.title || '',
                    order: index + 1,
                };
            });

            const payload = { ...values, elements };

            if (id) {
                await updateReportDefinition(id, payload);
                message.success('報表定義已成功更新！');
            } else {
                await createReportDefinition(payload);
                message.success('報表定義已成功新增！');
            }
            navigate('/reports');

        } catch (info) {
            console.log('Validate Failed:', info);
        } finally {
            setIsSaving(false);
        }
    };

    const moveItem = useCallback((dragIndex: number, hoverIndex: number) => {
        const dragItemKey = targetKeys[dragIndex];
        setTargetKeys(
            update(targetKeys, {
                $splice: [[dragIndex, 1], [hoverIndex, 0, dragItemKey]],
            }),
        );
    }, [targetKeys]);

    const transferDataSource = availableElements.map(el => ({ ...el, key: el.id }));

    return (
        <DndProvider backend={HTML5Backend}>
            <Title level={2}>{id ? '編輯報表定義' : '新增報表定義'}</Title>
            <Form form={form} layout="vertical" onFinish={onSave} disabled={loading}>
                <Title level={4}>基本資訊</Title>
                <Form.Item label="報表名稱" name="name" rules={[{ required: true, message: '請輸入報表名稱' }]}>
                    <Input placeholder="例如：每日網站流量分析報表" />
                </Form.Item>
                <Form.Item label="選擇資料來源" name="datasource_id" rules={[{ required: true, message: '請選擇一個資料來源' }]}>
                    <Select placeholder="選擇一個已驗證的資料來源" disabled={!!id}>
                        {dataSources.map(ds => (
                            <Option key={ds.id} value={ds.id} disabled={ds.status !== 'verified'}>
                                {ds.name} ({ds.type.toUpperCase()})
                            </Option>
                        ))}
                    </Select>
                </Form.Item>
                <Form.Item label="時間範圍">
                    <Space.Compact>
                        <Form.Item name={['time_range_quick']} noStyle>
                            <Select placeholder="選擇快捷時間" style={{ width: '150px' }} onChange={(value) => form.setFieldsValue({ time_range: value, time_range_custom_value: null, time_range_custom_unit: 'h' })}>
                                <Option value="now-1h">過去 1 小時</Option>
                                <Option value="now-24h">過去 24 小時</Option>
                                <Option value="now-7d">過去 7 天</Option>
                                <Option value="now-30d">過去 30 天</Option>
                            </Select>
                        </Form.Item>
                         <Form.Item name={['time_range_custom_value']} noStyle>
                            <Input
                                style={{ width: '150px' }}
                                placeholder="或自訂相對時間"
                                type="number"
                                onChange={(e) => form.setFieldsValue({ time_range_quick: null, time_range: `now-${e.target.value}${form.getFieldValue('time_range_custom_unit')}`})}
                            />
                        </Form.Item>
                        <Form.Item name={['time_range_custom_unit']} noStyle>
                             <Select style={{ width: '80px' }} onChange={(value) => form.setFieldsValue({ time_range_quick: null, time_range: `now-${form.getFieldValue('time_range_custom_value')}${value}`})}>
                                <Option value="m">分鐘</Option>
                                <Option value="h">小時</Option>
                                <Option value="d">天</Option>
                                <Option value="w">週</Option>
                                <Option value="M">月</Option>
                                <Option value="y">年</Option>
                            </Select>
                        </Form.Item>
                    </Space.Compact>
                </Form.Item>
                {/* 真正提交到後端的欄位，使用者不可見 */}
                <Form.Item name="time_range" hidden>
                    <Input />
                </Form.Item>
                <Divider />
                <Title level={4}>挑選報表元素</Title>
                <p>請從左側選擇您想包含在此報表中的儀表板或圖表，並可拖曳右側項目進行排序。</p>
                <Transfer
                    dataSource={transferDataSource}
                    titles={['可選項目', '已選項目']}
                    targetKeys={targetKeys}
                    onChange={(newTargetKeys) => setTargetKeys(newTargetKeys as string[])}
                    render={item => item.title}
                    listStyle={{ width: '100%', height: 300 }}
                >
                    {({ direction, filteredItems, onItemSelect, selectedKeys }) => {
                        if (direction === 'left') {
                            return (
                                <Spin spinning={elementsLoading} tip="讀取中...">
                                    <Table
                                        rowSelection={{ onSelectAll(selected, selectedRows) { const treeSelectedKeys = selectedRows.map(({ key }) => key); const diffKeys = selected ? treeSelectedKeys.filter(key => !selectedKeys.includes(String(key))) : treeSelectedKeys; onItemSelect(diffKeys as any, selected); }, onSelect({ key }, selected) { onItemSelect(String(key), selected); }, selectedRowKeys: selectedKeys as any }}
                                        columns={[{ dataIndex: 'title', title: '名稱' }]}
                                        dataSource={filteredItems}
                                        size="small"
                                        onRow={({ key }) => ({ onClick: () => { onItemSelect(String(key), !selectedKeys.includes(String(key))); } })}
                                    />
                                </Spin>
                            );
                        }
                        return (
                            <div style={{ height: '100%', overflow: 'auto' }}>
                                {targetKeys.map((key, index) => {
                                    const item = availableElements.find(i => i.id === key);
                                    return item ? <DraggableItem key={key} index={index} id={key} text={item.title} moveItem={moveItem} /> : null;
                                })}
                            </div>
                        );
                    }}
                </Transfer>
                <Divider />
                <Form.Item>
                    <Button type="primary" htmlType="submit" loading={isSaving}>儲存</Button>
                    <Button style={{ marginLeft: 8 }} onClick={() => navigate('/reports')}>取消</Button>
                </Form.Item>
            </Form>
        </DndProvider>
    );
};

export default ReportDefinitionForm;
