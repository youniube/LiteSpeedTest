<template>
  <div>
    <el-row>
        <el-col :span="22" :offset="1">
            <el-card>
                        <div slot="header">订阅信息</div>
                        <el-container :element-loading-text="loadingContent">
                            <el-form label-width="120px" size="large">
                                    <el-form-item label="设置：">
                                        <el-radio-group v-model="option" :disabled="loading">
                                            <el-radio :label="0">基础</el-radio>
                                            <el-radio :label="1">高级</el-radio>
                                            <el-radio :label="2">导出</el-radio>
                                            <el-radio :label="3">手动生成</el-radio>
                                        </el-radio-group>
                                    </el-form-item>

                                    <el-form-item label="链接：" v-if="option<2">
                                        <el-col :span="8">
                                            <el-input v-model="subscription" style="width: 800px"
                                            @keyup.enter.native="submit" placeholder="支持 VLESS/V2Ray/Trojan/SS/SSR/Clash 订阅链接，订阅文件及节点链接的批量测速"
                                            :disabled="loading||upload" clearable></el-input>
                                        </el-col>
                                        <el-col :span="12" ></el-col>
                                        <el-col :span="12" style="margin-top: 6px;">
                                            <el-upload :drag="checkUploadStatus('drag')" :v-if="checkUploadStatus('if')"
                                            action="" :show-file-list="false" ref="upload" :http-request="handleFileChange"
                                            :auto-upload="true" :before-upload="beforeUpload">
                                            <i class="el-icon-upload" v-if="!subscription.length"></i>
                                            <!--<el-button slot="trigger" type="primary" icon="el-icon-files" :disabled="loading" v-if="!upload">选择配置文件</el-button>-->
                                            <el-button slot="tip" type="danger" icon="el-icon-close" :disabled="loading"
                                                v-if="upload" @click="cancelFileUpload">取消文件选择</el-button>
                                            <div class="el-upload__text" v-if="!subscription.length">
                                                还可以将配置文件拖到此处，或<em>点击上传</em></div>
                                            </el-upload>
                                        </el-col>
                                        
                                    </el-form-item>

                                    <el-form-item label="并发数：" v-if="option<2">
                                        <el-input v-model="concurrency" style="width: 235px" type="number" min="1" max="50"
                                            @keyup.enter.native="submit" :disabled="loading"></el-input>
                                    </el-form-item>
                                    <el-form-item label="测试时长(秒)：" v-if="option===1">
                                        <el-input v-model="timeout" style="width: 235px" type="number" min="5" max="60"
                                            @keyup.enter.native="submit" :disabled="loading"></el-input>
                                    </el-form-item>
                                    <el-form-item label="去除重复节点：" v-if="option===1">
                                        <el-checkbox v-model="unique">去重</el-checkbox>
                                    </el-form-item>
                                    <el-form-item label="测速引擎：" v-if="option===1">
                                        <el-select v-model="engine" :disabled="loading">
                                            <el-option v-for="(item, key, index) in init.engines" :key="index"
                                                :label="item" :value="key">
                                            </el-option>
                                        </el-select>
                                    </el-form-item>
                                    <el-form-item label="sing-box程序：" v-if="option===1 && engine !== 'native'">
                                        <el-input v-model="singboxBin" style="width: 235px"
                                            @keyup.enter.native="submit" :disabled="loading" clearable></el-input>
                                    </el-form-item>
                                    <el-form-item label="临时目录：" v-if="option===1 && engine !== 'native'">
                                        <el-input v-model="singboxWorkDir" style="width: 235px"
                                            @keyup.enter.native="submit" :disabled="loading" clearable></el-input>
                                    </el-form-item>
                                    <el-form-item label="保留临时文件：" v-if="option===1 && engine !== 'native'">
                                        <el-checkbox v-model="keepTempFile">保留</el-checkbox>
                                    </el-form-item>
                                    <el-form-item label="测试项：" v-if="option<2">
                                        <el-select v-model="speedtestMode" :disabled="loading">
                                            <el-option v-for="(item, key, index) in init.speedtestModes" :key="index"
                                                :label="item" :value="key">
                                            </el-option>
                                        </el-select>
                                    </el-form-item>
                                    <el-form-item label="重命名：" v-if="option<2">
                                        <el-checkbox v-model="renameUseExternal" :disabled="loading">外部查询</el-checkbox>
                                        <el-input v-model="renameIntervalMs" style="width: 160px; margin-left: 12px" type="number" min="300" max="5000"
                                            :disabled="loading || !renameUseExternal">
                                            <template #append>ms</template>
                                        </el-input>
                                    </el-form-item>
                                    <el-form-item label="自定义组名：" v-if="option<2">
                                        <el-input v-model="groupname" style="width: 235px" 
                                            @keyup.enter.native="submit" :disabled="loading" clearable></el-input>
                                        <el-button type="primary" @click="submit" style="margin-left: 20px" v-if="!option"
                                            :disabled="loading" :loading="loading"><el-icon v-if="!loading" class="el-icon--left"><Select /></el-icon>提 交</el-button>
                                        <el-button type="danger" @click="terminate" :icon="Close" v-if="!option"
                                            :disabled="!loading"><el-icon class="el-icon--left"><Close /></el-icon>终 止</el-button>
                                    </el-form-item>

                                    <el-form-item label="Ping方式：" v-if="option===1" :disabled="loading">
                                        <el-select v-model="pingMethod" >
                                            <el-option v-for="(item, key, index) in init.pingMethods" :key="index" :label="item" :value="key">
                                            </el-option>
                                        </el-select>
                                        <el-button type="primary" @click="submit" style="margin-left: 20px" :icon="Check" :disabled="loading"
                                            :loading="loading">提 交</el-button>
                                        <el-button type="danger" @click="terminate" :icon="Close" :disabled="!loading">终 止</el-button>
                                    </el-form-item>
                                    
                                    <!-- export -->
                                    <el-form-item label="语言：" v-if="option===2" :disabled="loading">
                                        <el-select v-model="language">
                                            <el-option key="1" label="EN" value="en"></el-option>
                                            <el-option key="2" label="中文" value="cn"></el-option>
                                        </el-select>
                                    </el-form-item>
                                    <el-form-item label="字体大小：" v-if="option===2">
                                        <el-input v-model="fontSize" style="width: 235px" type="number" min="12" max="36"
                                        @keyup.enter.native="submit" :disabled="loading"></el-input>
                                    </el-form-item>
                                    <el-form-item label="排序方式：" v-if="option===2" :disabled="loading">
                                        <el-select v-model="sortMethod" >
                                            <el-option v-for="(item, key, index) in init.sortMethods" :key="index"
                                                :label="item" :value="key">
                                            </el-option>
                                        </el-select>
                                    </el-form-item>
                                    <el-form-item label="主题：" v-if="option===2" :disabled="loading">
                                        <el-select v-model="theme" >
                                            <el-option v-for="(item, key, index) in init.themes" :key="index"
                                                :label="item" :value="key">
                                            </el-option>
                                        </el-select>
                                    </el-form-item>
                                    <el-form-item label="结果数据：" v-if="option===3" :disabled="loading">
                                        <el-input
                                            type="textarea"
                                            :autosize="{ minRows: 5, maxRows: 18}"
                                            placeholder="input data"
                                            style="width: 800px"
                                            v-model="generateResultJSON">
                                        </el-input>
                                        <el-button type="primary" @click="generateResult" style="margin-left: 20px" :icon="Check" :disabled="loading"
                                            :loading="loading">生 成</el-button>
                                    </el-form-item>
                            </el-form>
                        </el-container>
                    </el-card>

            <br>

                    <el-card>
                        <el-row style="display: flex;">
                            <el-col style="display: flex;align-items: center;" :span="1">
                                <div>结果</div>
                            </el-col>
                            <el-col v-if="result.length" :span="8">
                                <el-dropdown trigger="click">
                                    <el-button size="large" type="primary">
                                        Actions<el-icon class="el-icon--right"><arrow-down /></el-icon>
                                    </el-button>
                                    <template #dropdown>
                                        <el-dropdown-menu slot="dropdown">
                                            <el-dropdown-item @click.native="handleCopySub()">复制订阅链接</el-dropdown-item>
                                            <el-dropdown-item v-if="!loading && result.length" @click.native="handleCopyAvailable()">复制可用节点</el-dropdown-item>
                                            <el-dropdown-item v-if="!loading && result.length" @click.native="handleSmartRename()">重命名节点</el-dropdown-item>
                                            <el-dropdown-item v-if="selectedNodes.length" @click.native="handleCopy()">复制节点</el-dropdown-item>
                                            <el-dropdown-item v-if="selectedNodes.length" @click.native="handleSave()">导出节点</el-dropdown-item>
                                            <!-- <el-dropdown-item @click.native="handleRetest()">重新测试</el-dropdown-item> -->
                                            <el-dropdown-item v-if="selectedNodes.length" @click.native="handleQRCode()">显示二维码</el-dropdown-item>
                                            <el-dropdown-item v-if="selectedNodes.length" @click.native="handleExportResult()">导出结果</el-dropdown-item>
                                        </el-dropdown-menu>
                                    </template>
                                </el-dropdown>
                            </el-col>
                        </el-row>
                        <el-container>
                            <!-- https://www.ag-grid.com/vue-data-grid/grid-size/#grid-auto-height  max 300row -->
                                <ag-grid-vue 
                                        id="myGrid"
                                        style="width: 100%;padding-top: 12px;" 
                                        class="ag-theme-alpine"
                                        :domLayout="domLayout"
                                        :rowData="result"
                                        :columnDefs="columns"
                                        :getRowId="getRowId"
                                        :rowSelection="rowSelection"
                                        :suppressRowClickSelection="true"
                                        :defaultColDef="defaultColDef"
                                        @grid-ready="onGridReady"
                                        @selection-changed="onSelectionChanged"
                                        @sort-changed="scheduleDerivedStateSync"
                                        @filter-changed="scheduleDerivedStateSync"
                                    >
                                </ag-grid-vue>
                        </el-container>
                        <!-- <el-container>
                            <el-table :data="result" :cell-style="colorCell" ref="result" 
                                :row-key="row => `${row.server}${row.protocol}${row.ping}${row.speed}${row.maxspeed}`"
                                @selection-change="handleSelectionChange" @sort-change="handleSortChange">
                                <el-table-column type="selection" width="55" :selectable="checkSelectable">
                                </el-table-column>
                                <el-table-column label="Remark" align="center" prop="remark" min-width="400" sortable>
                                </el-table-column>
                                <el-table-column label="Server" align="center" prop="server" min-width="160" sortable>
                                </el-table-column>
                                <el-table-column label="Protocol" align="center" prop="protocol" width="120" sortable
                                    :filters="[{ text: 'V2Ray', value: 'vmess' }, { text: 'Trojan', value: 'trojan' }, { text: 'ShadowsocksR', value: 'ssr' }, { text: 'Shadowsocks', value: 'ss' }]"
                                    :filter-method="filterProtocol">
                                </el-table-column>
                                <el-table-column label="Ping" align="center" prop="ping" width="100" sortable="custom"
                                    :filters="[{ text: 'Available ', value: 'available' }]"
                                    :filter-method="filterPing">
                                </el-table-column>
                                <el-table-column label="AvgSpeed" align="center" prop="speed" min-width="150" sortable
                                    :filters="[{ text: '200KB', value: 204800 }, { text: '500KB', value: 512000 }, { text: '1MB', value: 1048576 }, { text: '2MB', value: 2097152 }, { text: '5MB', value: 5242880 }, { text: '10MB', value: 10485760 },{ text: '15MB', value: 15728640 }, { text: '20MB', value: 20971520 }]"
                                    :filter-multiple="false"
                                    :filter-method="filterAvgSpeed"
                                    :sort-method="speedSort">
                                </el-table-column>
                                <el-table-column label="MaxSpeed" align="center" prop="maxspeed" min-width="150" sortable
                                    :filters="[{ text: '200KB', value: 204800 }, { text: '500KB', value: 512000 }, { text: '1MB', value: 1048576 }, { text: '2MB', value: 2097152 }, { text: '5MB', value: 5242880 }, { text: '10MB', value: 10485760 },{ text: '15MB', value: 15728640 }, { text: '20MB', value: 20971520 }]"
                                    :filter-multiple="false"
                                    :filter-method="filterMaxSpeed"
                                    :sort-method="maxSpeedSort">
                                </el-table-column>
                            </el-table>
                        </el-container> -->
                    </el-card>   

            <br>

                        <div :class="['dashboard', dashboardCollapsed ? 'collapsed' : '']">
                        <el-card class="progress">
                            <div class="progress-bar" :style="{ 'width': testProgress(result, testCount) + '%' }"></div>
                            <div class="progress-inner">
                                <div class="progress-item">
                                    <span>{{ testProgress(result, testCount) }}%</span>
                                    <div>Progress</div>
                                </div>
                                <div v-if="!dashboardCollapsed" class="progress-item">
                                    <span>{{ availableCount(result) }}/{{ result.length }}</span>
                                    <div>Ratio</div>
                                </div>
                                <div v-if="!dashboardCollapsed" class="traffic">
                                    <span> {{ bytesToSize(totalTraffic) }} </span>
                                    <div>Traffic</div>
                                </div>
                                <div v-if="!dashboardCollapsed" class="time">
                                    <span>{{ formatSeconds(totalTime) }}</span>
                                    <div>Time</div>
                                </div>
                            </div>
                        </el-card>
                        <el-card class="category" v-memo="[result]">
                            <ul>
                                <li v-if="dashboardCollapsed">
                                    <span>{{ availableCount(result) }}/{{ result.length }}</span>
                                    <div>Ratio</div>
                                </li>
                                <li v-if="!dashboardCollapsed">
                                    <span>{{ result.filter(item => item.protocol.startsWith("vmess")).length }}</span>
                                    <div>Vmess</div>
                                </li>
                                <li v-if="!dashboardCollapsed">
                                    <span>{{ result.filter(item => item.protocol.startsWith("vless")).length }}</span>
                                    <div>Vless</div>
                                </li>
                                <li v-if="!dashboardCollapsed">
                                    <span>{{ result.filter(item => item.protocol === "trojan" || item.protocol.startsWith("trojan/")).length }}</span>
                                    <div>Trojan</div>
                                </li>
                                <li v-if="!dashboardCollapsed">
                                    <span>{{ result.filter(item => item.protocol === "ssr").length }}</span>
                                    <div>SSR</div>
                                </li>
                                <li v-if="!dashboardCollapsed">
                                    <span>{{ result.filter(item => item.protocol === "ss").length }}</span>
                                    <div>SS</div>
                                </li>
                            </ul>
                        </el-card>
                        <div class="icon" @click="handleDashboardCollapsed()">
                            <el-icon v-if="!dashboardCollapsed"><Right /></el-icon>
                            <el-icon v-if="dashboardCollapsed"><Back /></el-icon>
                            <!-- <i v-if="!dashboardCollapsed" class="el-icon-right" ></i>
                            <i v-if="dashboardCollapsed" class="el-icon-back"></i> -->
                        </div>
                    </div>

            <br>

            <el-card v-if="picdata.length">
                <div slot="header">导出图片</div>
                <el-container>
                    <el-image :src="picdata" fit="true" placeholder="未加载图片" id="result_png"></el-image>
                </el-container>
            </el-card>
        </el-col>
    </el-row>
    <el-dialog title="Share Links with QRcode" center v-model="qrCodeDialogVisible" width="40%"
        @opened="handleQRCodeCreate" :before-close="qrCodeHandleClose">
        <el-scrollbar style="height:360px;">
            <el-row>
                <el-col v-for="(item, index) of selectedNodes" :key="index" :span="12">
                    <el-card :body-style="{ padding: '0px', height:'400px'}" shadow="hover"
                        style="width: 320px;height: 330px;text-align: center;">
                        <div style="display: flex; flex-direction: column; align-items: center; justify-content: center; margin-top: 15px;">
                            <div :id="'qrcode_' + item.id" style="margin-left: 20px;"></div>
                            <div class="truncate_remark">{{ item.remark }}</div>
                            <div>{{ `${item.ping}ms ${item.speed} ${item.maxspeed}` }}</div>
                        </div>
                    </el-card>
                </el-col>
            </el-row>
        </el-scrollbar>
    </el-dialog>
  </div>
</template>

<script>

import "ag-grid-community/dist/styles/ag-grid.css";
import "ag-grid-community/dist/styles/ag-theme-alpine.css";
import { AgGridVue } from 'ag-grid-vue3';

const go = new Go();
    WebAssembly.instantiateStreaming(fetch("main.wasm"), go.importObject).then((result) => {
      go.run(result.instance);
});

let themes = {
    "original": {
        colorgroup: [
            [255, 255, 255],
            [128, 255, 0],
            [255, 255, 0],
            [255, 128, 192],
            [255, 0, 0]
        ],
        bounds: [0, 64 * 1024, 512 * 1024, 4 * 1024 * 1024, 16 * 1024 * 1024],
    },
    "rainbow": {
        colorgroup: [
                [255, 255, 255],
                [102, 255, 102],
                [255, 255, 102],
                [255, 178, 102],
                [255, 102, 102],
                [226, 140, 255],
                [102, 204, 255],
                [102, 102, 255]
            ],
            bounds: [0, 64 * 1024, 512 * 1024, 4 * 1024 * 1024, 16 * 1024 * 1024, 24 * 1024 * 1024, 32 * 1024 * 1024, 40 * 1024 * 1024 ]
    } 
}

let ws = null
const API_ROUTES = Object.freeze({
    test: "/test",
    getSubscriptionLink: "/getSubscriptionLink",
    generateResult: "/generateResult",
    renameNodes: "/renameNodes",
})

export default {
    data() {
        return {
            upload: false,
            filecontent: "",
            loading: false,
            subscription: "",
            concurrency: 2,
            timeout: 15,
            unique: true,
            engine: "",
            singboxBin: "sing-box",
            singboxWorkDir: ".lite-singbox",
            keepTempFile: false,
            groupname: "",
            loadingContent: "",
            speedtestMode: "all",
            pingMethod: "googleping",
            sortMethod: "rspeed",
            exportMaxSpeed: true,
            method: "SOCKET",
            picdata: "",
            option: 0,
            multipleSelection: [],
            selectedNodeIds: [],
            qrCodeDialogVisible: false,
            totalTraffic: 0,
            totalTime: 0,
            language: "en",
            fontSize: 24,
            theme: "rainbow",
            generateResultJSON: "",
            dashboardCollapsed: true,
            testCount: 0,
            testOkCount: 0,
            renameUseExternal: true,
            renameIntervalMs: 1200,
            sortState: {},
            resultSyncTimer: null,
            // agGrid
            columns: this.columns,
            gridApi: null,
            getRowId: null,
            domLayout: null,
            rowSelection: null,
            defaultColDef: {
                resizable: true,
                sortable: true,
                cellStyle: { textAlign: 'center' },
            },

            init: {
                speedtestModes: {
                    all: "全部",
                    pingonly: "Ping only",
                    speedonly: "Speed only",
                },
                pingMethods: {
                    googleping: "Google",
                    tcping: "TCP",
                },
                engines: {
                    "": "自动（VLESS 使用 sing-box）",
                    native: "原生引擎",
                    singbox: "强制 sing-box",
                },
                sortMethods: {
                    rspeed: "speed 倒序",
                    speed: "speed 顺序",
                    ping: "ping 顺序",
                    rping: "ping 倒序",
                    none: "默认",
                },
                themes: {
                    rainbow: "Rainbow",
                    original: "Original",
                }
            },
            result: []
        }
    },
    components: {
        'ag-grid-vue': AgGridVue
    },
    computed: {
        selectedNodes() {
            return this.getSelectedNodes();
        },
    },
    watch: {
        result: {
            deep: true,
            handler() {
                this.scheduleDerivedStateSync();
            },
        },
        totalTraffic() {
            this.scheduleDerivedStateSync();
        },
        totalTime() {
            this.scheduleDerivedStateSync();
        },
        language() {
            this.scheduleDerivedStateSync();
        },
        fontSize() {
            this.scheduleDerivedStateSync();
        },
        theme() {
            this.scheduleDerivedStateSync();
        },
        sortMethod() {
            this.scheduleDerivedStateSync();
        },
    },
    beforeUnmount() {
        if (this.resultSyncTimer) {
            clearTimeout(this.resultSyncTimer);
            this.resultSyncTimer = null;
        }
        this.disconnect();
    },
    created() {        
        this.columns = Object.freeze([
                { headerName: 'Remark', field: 'remark', headerCheckboxSelection: true,checkboxSelection: true, minWidth: 500, flex: 1, filter: 'agTextColumnFilter', filterParams: {suppressAndOrCondition: true} },
                { headerName: 'Server', field: 'server', minWidth: 330, filter: 'agTextColumnFilter', filterParams: {suppressAndOrCondition: true} },
                { headerName: "Protocol", field: 'protocol', width: 150, filter: 'agTextColumnFilter' },
                { headerName: 'Ping(ms)', field: 'ping', width: 200, sortingOrder: ['desc', 'asc', null], comparator: (valueA, valueB, nodeA, nodeB, isInverted) => {
                    // isInverted: true for Ascending, false for Descending.
                    if (isInverted) {
                        let ping1 = parseFloat(valueB);
                        if (ping1 < 1) { ping1 = 99999 }
                        let ping2 = parseFloat(valueA);
                        if (ping2 < 1) { ping2 = 99999 }
                        return ping1 - ping2
                    } else  {
                        return parseFloat(valueB) - parseFloat(valueA)
                    } 
                }, filter: 'agNumberColumnFilter', filterParams: { 
                    suppressAndOrCondition: true,
                    filterOptions: [
                        { displayKey: "lessThanOrEqual", 
                        displayName: "<=", 
                        predicate: ([filterValue], cellValue) =>  cellValue > 0 && cellValue <= filterValue }
                    ] }},
                { headerName: 'AvgSpeed', field: 'speed', width: 200, cellStyle: this.speedCellStyle, sortingOrder: ['asc', null], comparator: (valueA, valueB, nodeA, nodeB, isInverted) => {
                        const speed1 = isNaN(this.getSpeed(valueA)) ? -1 : this.getSpeed(valueA);
                        const speed2 = isNaN(this.getSpeed(valueB)) ? -1 : this.getSpeed(valueB);
                        return speed2 - speed1
                }},
                { headerName: 'MaxSpeed', field: 'maxspeed', width: 200, cellStyle: this.speedCellStyle,sortingOrder: ['asc', null], comparator: (valueA, valueB, nodeA, nodeB, isInverted) => {
                        const speed1 = isNaN(this.getSpeed(valueA)) ? -1 : this.getSpeed(valueA);
                        const speed2 = isNaN(this.getSpeed(valueB)) ? -1 : this.getSpeed(valueB);
                        return speed2 - speed1
                }},
            ])
        this.getRowId = params => params.data.id;
        this.rowSelection = 'multiple';     
        this.domLayout = 'autoHeight';
    },
    methods: {
        updateRow(id, newData) {
            if (!this.gridApi) {
                return;
            }
            const rowNode = this.gridApi.getRowNode(`${id}`) || this.gridApi.getRowNode(id);
            if (!rowNode) {
                this.gridApi.applyTransaction({ add: [newData] })
                return;
            }
            rowNode.setData({ ...rowNode.data, ...newData });
            this.gridApi.refreshCells({ rowNodes: [rowNode], force: true });
        },
        updateRowAsync(id, newData) {
            if (!this.gridApi) {
                return;
            }
            const rowNode = this.gridApi.getRowNode(`${id}`) || this.gridApi.getRowNode(id);
            if (!rowNode) {
                this.gridApi.applyTransactionAsync({ add: [newData] })
                return;
            }
            rowNode.setData({ ...rowNode.data, ...newData });
            this.gridApi.refreshCells({ rowNodes: [rowNode], force: true });
        },
        setAutoHeight() {
            this.gridApi.setDomLayout('autoHeight');
            // auto height will get the grid to fill the height of the contents,
            // so the grid div should have no height set, the height is dynamic.
            document.querySelector('#myGrid').style.height = '';
        },
        setFixedHeight() {
            // we could also call setDomLayout() here as normal is the default
            this.gridApi.setDomLayout('normal');
            // when auto height is off, the grid ahs a fixed height, and then the grid
            // will provide scrollbars if the data does not fit into it.
            document.querySelector('#myGrid').style.height = '3000px';
        },
        speedCellStyle(params) {
            // console.log(`params.value: ${params.value}`)
            const style = {textAlign: 'center'}
            const speed = this.getSpeed(params.value);
            if (speed < 1 || isNaN(parseFloat(speed))) return style;
            const color = this.getSpeedColor(speed);
            style.backgroundColor = color;
            return { backgroundColor: color, textAlign: 'center' }
        },
        onGridReady(params) {
            this.gridApi = params.api;
            this.scheduleDerivedStateSync();
            // this.gridColumnApi = params.columnApi;
        },
        onSelectionChanged() {
            const selectedRows = this.gridApi ? this.gridApi.getSelectedRows() : [];
            this.selectedNodeIds = selectedRows.map(item => item.id);
            this.multipleSelection = this.getSelectedNodes();
        },
        getSelectedNodes() {
            if (!Array.isArray(this.selectedNodeIds) || !this.selectedNodeIds.length) {
                return [];
            }
            const selectedIds = new Set(this.selectedNodeIds);
            if (this.gridApi) {
                const items = [];
                this.gridApi.forEachNodeAfterFilterAndSort(node => {
                    if (node && node.data && selectedIds.has(node.data.id)) {
                        items.push(node.data);
                    }
                });
                return items;
            }
            return this.result.filter(item => item && selectedIds.has(item.id));
        },
        requireSelectedNodes(actionLabel) {
            const nodes = this.getSelectedNodes();
            if (!nodes.length) {
                this.$message.warning(`${actionLabel}前请先选择节点`);
                return null;
            }
            return nodes;
        },
        async readAPIErrorMessage(resp, fallback = 'Request failed') {
            const raw = await resp.text();
            if (!raw) {
                return fallback;
            }
            try {
                const data = JSON.parse(raw);
                return data.error || data.message || raw;
            } catch (_) {
                return raw;
            }
        },
        async requestJSON(path, options = {}) {
            const resp = await fetch(this.apiPath(path), options);
            if (!resp.ok) {
                throw new Error(await this.readAPIErrorMessage(resp));
            }
            return await resp.json();
        },
        async requestText(path, options = {}) {
            const resp = await fetch(this.apiPath(path), options);
            if (!resp.ok) {
                throw new Error(await this.readAPIErrorMessage(resp));
            }
            return await resp.text();
        },
        normalizeBase64(input) {
            const normalized = `${input || ""}`.replace(/-/g, '+').replace(/_/g, '/').replace(/\s+/g, '');
            const padding = normalized.length % 4;
            return padding ? normalized + '='.repeat(4 - padding) : normalized;
        },
        decodeBase64Unicode(input) {
            try {
                return decodeURIComponent(escape(window.atob(this.normalizeBase64(input))));
            } catch (_) {
                return window.atob(this.normalizeBase64(input));
            }
        },
        encodeBase64Unicode(input) {
            return window.btoa(unescape(encodeURIComponent(`${input || ''}`)));
        },
        apiPath(path) {
            return `${path}`;
        },
        wsURL(path) {
            const scheme = window.location.protocol === "https:" ? "wss" : "ws";
            return `${scheme}://${window.location.host}${path}`;
        },
        buildNodeLink(item) {
            if (!item) {
                return "";
            }
            return this.rewriteLinkRemark(item.link, item.remark);
        },
        replaceHashRemark(link, remark) {
            const trimmed = `${link || ''}`.trim();
            if (!trimmed || !remark) {
                return trimmed;
            }
            const base = trimmed.includes('#') ? trimmed.slice(0, trimmed.indexOf('#')) : trimmed;
            return `${base}#${encodeURIComponent(remark)}`;
        },
        rewriteVmessRemark(link, remark) {
            const trimmed = `${link || ''}`.trim();
            if (!trimmed || !remark) {
                return trimmed;
            }
            const match = trimmed.match(/^vmess:\/\/([^#\s]+)(?:#.*)?$/i);
            if (!match) {
                return this.replaceHashRemark(trimmed, remark);
            }
            const payload = match[1];
            if (payload.includes('@')) {
                return this.replaceHashRemark(trimmed, remark);
            }
            try {
                const raw = this.decodeBase64Unicode(payload);
                const cfg = JSON.parse(raw);
                cfg.ps = remark;
                const encoded = this.encodeBase64Unicode(JSON.stringify(cfg));
                return `vmess://${encoded}#${encodeURIComponent(remark)}`;
            } catch (_) {
                return this.replaceHashRemark(trimmed, remark);
            }
        },
        rewriteLinkRemark(link, remark) {
            const trimmed = `${link || ''}`.trim();
            const nextRemark = `${remark || ''}`.trim();
            if (!trimmed || !nextRemark) {
                return trimmed;
            }
            const scheme = trimmed.split('://', 1)[0].toLowerCase();
            switch (scheme) {
                case 'vmess':
                    return this.rewriteVmessRemark(trimmed, nextRemark);
                case 'vless':
                case 'trojan':
                case 'ss':
                case 'http':
                    return this.replaceHashRemark(trimmed, nextRemark);
                default:
                    return this.replaceHashRemark(trimmed, nextRemark);
            }
        },
        getResultItems() {
            return this.result.filter(item => !!item);
        },
        getRenderableResultItems() {
            if (this.gridApi) {
                const items = [];
                this.gridApi.forEachNodeAfterFilterAndSort(node => {
                    if (node && node.data) {
                        items.push(node.data);
                    }
                });
                return items;
            }
            return this.getResultItems();
        },
        syncSelectionWithResult() {
            if (!Array.isArray(this.selectedNodeIds) || !this.selectedNodeIds.length) {
                this.multipleSelection = [];
                return;
            }
            const validIds = new Set(this.getResultItems().map(item => item.id));
            this.selectedNodeIds = this.selectedNodeIds.filter(id => validIds.has(id));
            this.multipleSelection = this.getSelectedNodes();
            if (this.gridApi) {
                this.gridApi.forEachNode(node => {
                    if (!node || !node.data) {
                        return;
                    }
                    node.setSelected(this.selectedNodeIds.includes(node.data.id));
                });
            }
        },
        scheduleDerivedStateSync() {
            if (this.resultSyncTimer) {
                clearTimeout(this.resultSyncTimer);
            }
            this.resultSyncTimer = window.setTimeout(() => {
                this.resultSyncTimer = null;
                this.syncSelectionWithResult();
                this.syncGenerateResultJSONFromCurrentResult();
            }, 60);
        },
        syncGenerateResultJSONFromCurrentResult() {
            const items = this.getResultItems();
            if (!items.length) {
                this.generateResultJSON = '';
                return null;
            }
            const data = this.buildRenderResultPayload();
            this.generateResultJSON = JSON.stringify(data, null, 2);
            return data;
        },
        clearSelectionState() {
            this.selectedNodeIds = [];
            this.multipleSelection = [];
            if (this.gridApi) {
                this.gridApi.deselectAll();
            }
        },
        clearDerivedResultState() {
            if (this.resultSyncTimer) {
                clearTimeout(this.resultSyncTimer);
                this.resultSyncTimer = null;
            }
            this.picdata = '';
            this.generateResultJSON = '';
        },
        getItemTraffic(item) {
            if (!item) {
                return 0;
            }
            const value = Number(item.traffic || 0);
            return Number.isFinite(value) && value > 0 ? value : 0;
        },
        recomputeTotalTraffic(items = null) {
            const sourceItems = Array.isArray(items) ? items.filter(item => !!item) : this.getResultItems();
            return sourceItems.reduce((sum, item) => sum + this.getItemTraffic(item), 0);
        },
        syncTotalTrafficFromResult() {
            this.totalTraffic = this.recomputeTotalTraffic();
            return this.totalTraffic;
        },
        buildRenderResultPayload(items = null) {
            const sourceItems = Array.isArray(items) ? items.filter(item => !!item) : this.getRenderableResultItems();
            const totalTraffic = this.recomputeTotalTraffic(sourceItems);
            const nodes = sourceItems.map(item => {
                const avg_speed = Math.floor(this.getSpeed(item.speed)) || 0;
                const max_speed = Math.floor(this.getSpeed(item.maxspeed)) || 0;
                return {
                    id: item.id,
                    group: item.group,
                    remarks: item.remark,
                    protocol: item.protocol,
                    ping: `${item.ping}`,
                    avg_speed,
                    max_speed,
                    isok: this.nodeAvailable(item),
                };
            });
            return {
                totalTraffic: this.bytesToSize(totalTraffic),
                totalTime: this.formatSeconds(this.totalTime),
                language: this.language,
                fontSize: this.fontSize,
                theme: this.theme,
                sortMethod: this.sortMethod,
                nodes,
            };
        },
        async requestResultImage(payload) {
            this.picdata = await this.requestText(API_ROUTES.generateResult, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload),
            });
        },
        parseManualResultPayload() {
            const raw = `${this.generateResultJSON || ''}`.trim();
            if (!raw.length) {
                throw new Error('结果数据为空');
            }
            return JSON.parse(raw);
        },
        async refreshResultImageFromCurrentResult() {
            const data = this.syncGenerateResultJSONFromCurrentResult();
            if (!data || !data.nodes.length) {
                this.picdata = '';
                return;
            }
            await this.requestResultImage(data);
        },
        applyRenamedNodes(nodes) {
            if (!Array.isArray(nodes) || !nodes.length) {
                return;
            }
            nodes.forEach(node => {
                const current = this.result[node.id];
                if (!current) {
                    return;
                }
                const nextRemark = node.remark || current.remark;
                const next = {
                    ...current,
                    remark: nextRemark,
                    link: this.rewriteLinkRemark(node.link || current.link, nextRemark),
                };
                this.result[node.id] = next;
                this.updateRow(node.id, next);
            });
            this.syncSelectionWithResult();
            this.scheduleDerivedStateSync();
        },
        async handleSmartRename(showNotice = true) {
            if (!Array.isArray(this.result) || !this.result.length) {
                return;
            }
            const payload = {
                useExternal: !!this.renameUseExternal,
                intervalMs: parseInt(this.renameIntervalMs, 10) || 1200,
                nodes: this.result
                    .filter(item => !!item)
                    .map(item => ({
                        id: item.id,
                        remark: item.remark,
                        server: item.server,
                        protocol: item.protocol,
                        link: item.link,
                    })),
            };
            try {
                const data = await this.requestJSON(API_ROUTES.renameNodes, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify(payload),
                });
                this.applyRenamedNodes(data.nodes || []);
                await this.refreshResultImageFromCurrentResult();
                if (showNotice) {
                    this.$notify.success('节点重命名完成');
                }
            } catch (err) {
                if (showNotice) {
                    this.$notify.error(`节点重命名失败：${err}`);
                }
                throw err;
            }
        },
        bytesToSize: function (bytes) {
            const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
            if (!bytes || bytes === 0) return '0 B';
            const i = parseInt(Math.floor(Math.log(Math.abs(bytes)) / Math.log(1024)), 10);
            if (i === 0) return `${bytes} ${sizes[i]})`;
            return `${(bytes / (1024 ** i)).toFixed(1)} ${sizes[i]}`;
        },
        metricPositive(value) {
            if (value === null || value === undefined) {
                return false;
            }
            if (typeof value === 'number') {
                return value > 0;
            }
            const raw = `${value}`.trim();
            if (!raw.length || raw === '测试中...') {
                return false;
            }
            if (raw.endsWith('B')) {
                return this.getSpeed(raw) > 0;
            }
            const number = parseFloat(raw);
            return !isNaN(number) && number > 0;
        },
        nodeAvailable(item) {
            if (!item) {
                return false;
            }
            return this.metricPositive(item.ping) || this.metricPositive(item.speed) || this.metricPositive(item.maxspeed);
        },
        availableCount(items) {
            return Array.isArray(items) ? items.filter(item => this.nodeAvailable(item)).length : 0;
        },
        testProgress: function (result, testCount) {
            return result.length ? Math.min(100, Math.floor(testCount/result.length*100)) : 0
        },
        formatSeconds: function (seconds) {
            let totalTime = seconds > 0 ? seconds : 0
            const hours = Math.floor(totalTime / 3600);
            totalTime %= 3600;
            const minutes = Math.floor(totalTime / 60);
            const secs = totalTime % 60;
            let result = `${secs}s`
            result = minutes > 0 ? `${minutes}m ${result}` : result
            result = hours > 0 ? `${hours}h ${result}` : result
            return result
        },
        incrTotalTime: function () {
            if (this.totalTime >= 0 && this.loading) {
                this.$nextTick(() => {
                    setTimeout(() => {
                        this.totalTime++;
                        this.incrTotalTime()
                    }, 1000);
                })
            }
        },
        cancelFileUpload: function () {
            let self = this;
            this.file = null;
            this.filecontent = '';
            this.subscription = '';
            self.upload = false;
        },
        handleFileChange(e) {
            let self = this;
            this.file = e.file;
            this.errText = '';
            if (!this.file || !window.FileReader) return;
            let reader = new FileReader();
            reader.readAsText(this.file);
            reader.onloadend = function () {
                self.filecontent = this.result;
                self.subscription = self.file.name;
                self.upload = true;
            }
        },
        beforeUpload(file) {
            // const isType = file.type === 'application/json' || file.type === 'application/octet-stream'
            const fsize = file.size / 1024 / 1024 <= 10;
            // if (!isType) {
            // 	this.$message.error('选择的文件格式有误!');
            // }
            if (!fsize) {
                this.$message.error('上传的文件不能超过10MB!');
            }
            return fsize;
        },
        checkUploadStatus(type) {
            if (!this.upload) {
                if (this.subscription.length)
                    return false;
                else
                    return true;
            }
            else {
                if (type === "if")
                    return true;
                else if (type === "drag")
                    return false;
            }
        },
        submit: function () {
            this.testCount = 0;
            this.testOkCount = 0;
            if (!this.subscription.length) {
                this.$alert("请先输入链接或选择文件！", "错误", {
                    type: "error",
                });
            } else {
                // this.$refs.result.clearSelection();
                // this.$refs.result.clearFilter();
                // this.$refs.result.clearSort();
                this.clearSelectionState();
                this.setAutoHeight()
                this.loading = true;
                this.totalTraffic = 0;
                this.totalTime = 0;
                this.clearDerivedResultState();
                this.result = [];
                this.incrTotalTime()
                this.loadingContent = "等待后端响应……";
                this.starttest();
            }
        },
        generateResult: function () {
            Promise.resolve()
                .then(() => this.option === 3
                    ? this.requestResultImage(this.parseManualResultPayload())
                    : this.refreshResultImageFromCurrentResult())
                .catch(err => {
                    this.$message.error(`Generate result failed: ${err}`);
                });
        },
        terminate: function () {
            this.loading = false;
            this.loadingContent = "等待后端响应……";
            this.result = [];
            this.clearSelectionState();
            this.clearDerivedResultState();
            this.disconnect();
        },
        handleSelectionChange(val) {
            // console.log(`select: ${JSON.stringify(val)}`)
            this.selectedNodeIds = Array.isArray(val) ? val.map(item => item.id) : [];
            this.multipleSelection = this.getSelectedNodes();
        },
        handleSortChange(val) {
            if (val.prop === "ping") { 
                if (val.order === "ascending") {
                    this.result.sort((obj1, obj2) => {
                        let ping1 = parseFloat(obj1.ping);
                        if (ping1 < 1) { ping1 = 99999 }
                        let ping2 = parseFloat(obj2.ping);
                        if (ping2 < 1) { ping2 = 99999 }
                        return ping1 - ping2
                    })
                } else if (val.order === "descending") {
                    this.result.sort((obj1, obj2) => parseFloat(obj2.ping) - parseFloat(obj1.ping))
                } else {
                    this.result.sort((obj1, obj2) => obj1.id - obj2.id)
                }
             }
        },
        copyToClipboard: async function (data) {
            if (navigator.clipboard) {
                    await navigator.clipboard.writeText(data)
                } else {
                    let textArea = document.createElement("textarea");
                    textArea.value = data;
                    // make the textarea out of viewport
                    textArea.style.position = "fixed";
                    textArea.style.left = "-999999px";
                    textArea.style.top = "-999999px";
                    document.body.appendChild(textArea);
                    textArea.focus();
                    textArea.select();
                    document.execCommand('copy');
                    textArea.remove();
                }
        },
        handleCopySub: async function () {
            try {
                const groupname = this.groupname.trim() || "Default";
                const exportNodes = this.getRenderableResultItems();
                const payload = { group: groupname };
                if (exportNodes.length) {
                    payload.links = exportNodes.map(item => this.buildNodeLink(item)).filter(link => !!link);
                }
                if (!payload.links || !payload.links.length) {
                    payload.filePath = this.subscription.trim();
                }
                const data = await this.requestJSON(API_ROUTES.getSubscriptionLink, {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify(payload)
                });
                await this.copyToClipboard(data.link || "");
                this.$message.success(exportNodes.length ? "Copy current subscription succeed!" : "Copy Subscription succeed!");
            } catch (err) {
                this.$message.error(`Copy Subscription failed: ${err}`);
            }
        },
        handleCopy: async function () {
            const selectedNodes = this.requireSelectedNodes('复制');
            if (!selectedNodes) {
                return;
            }
            try {
                const links = selectedNodes.map(elem => this.buildNodeLink(elem)).join("\n");
                await this.copyToClipboard(links);
                this.$message.success("Copy link succeed!");
            } catch (err) {
                this.$message.error("Copy link failed!");
            }
        },
        handleCopyAvailable: async function () {
            try {
                const links = this.getRenderableResultItems().filter(elem => this.nodeAvailable(elem)).map(elem => this.buildNodeLink(elem));
                if (!links.length) {
                    this.$message.warning('当前可见结果里没有可用节点');
                    return;
                }
                await this.copyToClipboard(links.join("\n"));
                this.$message.success(`Copy ${links.length} link${links.length>1 ? "s" : ""} succeed!`);
            } catch (err) {
                this.$message.error(`Copy link failed!`);
            }
        },
        qrCodeHandleClose() {
            this.qrCodeDialogVisible = false;
            this.getSelectedNodes().forEach(item => {
                const el = document.getElementById('qrcode_' + item.id);
                if (el) {
                    el.innerHTML = '';
                }
            });
        },
        handleQRCode() {
            this.qrCodeDialogVisible = true
        },
        handleQRCodeCreate: function () {
            const selectedNodes = this.requireSelectedNodes('生成二维码');
            if (!selectedNodes) {
                this.qrCodeDialogVisible = false;
                return;
            }
            this.$nextTick(() => {
                const items = selectedNodes.map(item => {
                    return {
                        gid: 'qrcode_' + item.id,
                        link: this.buildNodeLink(item),
                        size: 260
                    }
                })
                wasmQRcode(JSON.stringify(items))
                // this.multipleSelection.forEach(item => {
                // 	const gid = 'qrcode_' + item.id;
                // 	wasmQRcode(gid, item.link, 260, 260)
                // })
            })
        },
        handleRetest: function () {
            const selectedNodes = this.requireSelectedNodes('重新测试');
            if (!selectedNodes) {
                return;
            }
            const testids = selectedNodes.map(elem => elem.id);
            const links = selectedNodes.map(elem => this.buildNodeLink(elem));
            const data = { testMode: 3, ...this.getJSONOptions(), testids, links };
            console.log(`handleRetest: ${JSON.stringify(data)}`);
            this.send(JSON.stringify(data));
        },
        saveData: function (data, name) {
            const blob = new Blob([data], { type: 'text/plain;charset=utf-8;' })
            const link = document.createElement('a')
            if (link == null || link.download == null || link.download == undefined) {
                return
            }
            var event = new Date();
            event.setMinutes(event.getMinutes() - event.getTimezoneOffset());
            let jsonDate = event.toJSON().slice(0, 19);
            jsonDate = jsonDate.replaceAll("-", "")
            jsonDate = jsonDate.replaceAll("T", "")
            jsonDate = jsonDate.replaceAll(":", "")
            let url = URL.createObjectURL(blob)
            link.setAttribute('href', url)
            link.setAttribute('download', `${name}_${jsonDate}`)
            link.style.visibility = 'hidden'
            document.body.appendChild(link)
            link.click()
            document.body.removeChild(link)
            
        },
        handleDashboardCollapsed: function () {
            console.log(this.dashboardCollapsed)
            this.dashboardCollapsed = !this.dashboardCollapsed
        },
        handleSave: function () {
            const links = this.getSelectedNodes().map(elem => {
                    const link = this.rewriteLinkRemark(elem.link, elem.remark);
                    return `# ${elem.remark}\t${elem.ping}\t${elem.speed}\t${elem.maxspeed}\n${link}`
            })
            if (this.subscription.match(/^https?:\/\//g)) {
                links.unshift(`# ${this.subscription}`)
            }
            this.saveData(links.join("\n"), "profile")
        },
        handleExportResult: function (params) {
            const data = this.buildRenderResultPayload();
            this.saveData(JSON.stringify(data, null, 2), "result")
        },
        colorCell: function ({
            row,
            column,
            rowIndex,
            columnIndex
        }) {
            let style = {color: "black", "font-weight": 600};
            let speed = 0;
            switch (columnIndex) {
                case 5:
                    speed = this.getSpeed(row.speed);
                    break;
                case 6:
                    speed = this.getSpeed(row.maxspeed);
                    break;
                default:
                    return style;
            }
            if (isNaN(parseFloat(speed))) return style;
            let color = this.getSpeedColor(speed);
            // console.log(`speed: ${speed}, row.speed: ${row.speed}, row.maxspeed: ${row.maxspeed}  color: ${color}`);
            style.background = color
            return style;
        },
        // useNewPalette() {
        // 	colorgroup = [
        // 		[255, 255, 255],
        // 		[102, 255, 102],
        // 		[255, 255, 102],
        // 		[255, 178, 102],
        // 		[255, 102, 102],
        // 		[226, 140, 255],
        // 		[102, 204, 255],
        // 		[102, 102, 255]
        // 	];
        // 	bounds = [
        // 		0,
        // 		64 * 1024,
        // 		512 * 1024,
        // 		4 * 1024 * 1024,
        // 		16 * 1024 * 1024,
        // 		24 * 1024 * 1024,
        // 		32 * 1024 * 1024,
        // 		40 * 1024 * 1024
        // 	];
        // },
        getSpeed(speed) {
            let value = parseFloat(speed.toString().slice(0, -2));
            if (speed.toString().slice(-2) == "MB") {
                value *= 1048576;
            } else if (speed.toString().slice(-2) == "KB") {
                value *= 1024;
            } else value = parseFloat(speed.toString().slice(0, -1));
            return value;
        },
        getColor(lc, rc, level) {
            let colors = [];
            let r, g, b;
            colors.push(parseInt(lc[0] * (1 - level) + rc[0] * level));
            colors.push(parseInt(lc[1] * (1 - level) + rc[1] * level));
            colors.push(parseInt(lc[2] * (1 - level) + rc[2] * level));
            return colors;
        },
        getSpeedColor(speed) {
            const {colorgroup, bounds} = themes[this.theme];
            for (let i = 0; i < bounds.length - 1; i++) {
                if (speed >= bounds[i] && speed <= bounds[i + 1]) {
                    let color = this.getColor(
                        colorgroup[i],
                        colorgroup[i + 1],
                        (speed - bounds[i]) / (bounds[i + 1] - bounds[i])
                    );
                    return "rgb(" + color[0] + "," + color[1] + "," + color[2] + ")";
                }
            }
            return (
                "rgb(" +
                colorgroup[colorgroup.length - 1][0] +
                "," +
                colorgroup[colorgroup.length - 1][1] +
                "," +
                colorgroup[colorgroup.length - 1][2] +
                ")"
            );
        },
        connect(url) {
            try {
                ws = new WebSocket(url);
            } catch (ex) {
                this.loading = false;
                //this.$message.error('Cannot connect: ' + ex)
                this.$alert("后端连接错误！请检查后端运行情况！原因：" + ex, "错误");
                return;
            }
        },
        disconnect() {
            if (ws) {
                ws.close();
            }
        },
        send(msg) {
            if (ws) {
                try {
                    ws.send(msg);
                } catch (ex) {
                    this.$message.error("Cannot send: " + ex);
                }
            } else {
                this.loading = false;
                //this.$message.error('Cannot send: Not connected')
                this.$alert("后端连接错误！请检查后端运行情况！", "错误");
            }
        },
        getJSONOptions() {
            let self = this;
            let groupstr = self.groupname == "" ? "?empty?" : self.groupname;
            // const options = `^${groupstr}^${self.speedtestMode}^${self.pingMethod}^${self.sortMethod}^${self.exportMaxSpeed}^${self.concurrency}^${self.timeout}`
            return {
                group: groupstr,
                speedtestMode: self.speedtestMode,
                pingMethod: self.pingMethod,
                sortMethod: self.sortMethod,
                unique: self.unique,
                concurrency: parseInt(self.concurrency),
                timeout: parseInt(self.timeout),
                language: self.language,
                fontSize: parseInt(self.fontSize),
                theme: self.theme,
                engine: self.engine,
                singboxBin: self.singboxBin || "sing-box",
                singboxWorkDir: self.singboxWorkDir || ".lite-singbox",
                keepTempFile: !!self.keepTempFile,
            }
        },
        getOptions() {
            let self = this;
            let groupstr = self.groupname == "" ? "?empty?" : self.groupname;
            const options = `^${groupstr}^${self.speedtestMode}^${self.pingMethod}^${self.sortMethod}^${self.exportMaxSpeed}^${self.concurrency}^${self.timeout}`
            return options
        },
        starttest() {
            let self = this;
            let groupstr = self.groupname == "" ? "?empty?" : self.groupname;
            this.result = [];
            this.connect(this.wsURL(API_ROUTES.test));
            if (ws) {
                ws.addEventListener("open", function (ev) {
                    const data = self.getJSONOptions()
                    data.testMode = 2
                    data.subscription = self.upload ? self.filecontent : self.subscription;
                    this.send(JSON.stringify(data));
                });
                ws.addEventListener("message", this.MessageEvent);
            } else {
                this.loading = false;
                this.$alert("后端连接错误！请检查后端运行情况！", "错误");
            }
        },
        loopevent(id, tester) {
            const item = this.result[id];
            switch (tester) {
                case "ping":
                    item.ping = "测试中...";
                    item.loss = "测试中...";
                    item.testing = true
                    this.result[id]=item;
                    this.updateRow(id, item);
                    break;
                case "speed":
                    item.speed = "测试中...";
                    item.maxspeed = "测试中...";
                    item.testing = true
                    this.result[id]=item;
                    this.updateRow(id, item);
                    break;
            }
        },
        MessageEvent(ev) {
            console.log(ev.data);
            let json = JSON.parse(ev.data);
            let id = parseInt(json.id);

            let item = {};
            switch (json.info) {
                case "started":
                    this.loadingContent = "后端已启动……";
                    break;
                case "fetchingsub":
                    this.loadingContent = "正在获取节点，若节点较多将需要一些时间……";
                    break;
                case "begintest":
                    this.loadingContent = "疯狂测速中……";
                    break;
                case "gotserver":
                    item = {
                        id: id,
                        group: this.groupname == "" ? json.group : this.groupname,
                        remark: json.remarks,
                        server: json.server,
                        protocol: json.protocol,
                        link: json.link,
                        loss: "0.00%",
                        ping: "0.00",
                        speed: "0.00B",
                        maxspeed: "0.00B",
                        traffic: 0,
                        completed: false,
                    };
                    this.result[id] = item;
                    this.updateRow(id, item);
                    break;
                case "gotservers":
                    const items = json.servers.map(json => {
                        item = {
                            id: json.id,
                            group: this.groupname == "" ? json.group : this.groupname,
                            remark: json.remarks,
                            server: json.server,
                            protocol: json.protocol,
                            link: json.link,
                            loss: "0.00%",
                            ping: "0.00",
                            speed: "0.00B",
                            maxspeed: "0.00B",
                            completed: false,
                        };
                        this.result[json.id] = item;
                        return item
                    });
                    this.gridApi.applyTransaction({ add: items })
                    if (this.domLayout === "autoHeight" && this.result.length > 150) {
                        this.setFixedHeight()
                        this.domLayout = "normal"
                    } 
                    break;							
                case "endone":
                    item = this.result[id];
                    if (!item) {
                        break;
                    }
                    if (!item.completed) {
                        this.testCount += 1
                    }
                    item.completed = true
                    item.testing = false
                    this.result[id] = item;
                    this.updateRow(id, item);
                    break;
                case "startping":
                    //inverval=setInterval("app.loopevent("+id+",\"ping\")",300)
                    this.loopevent(id, "ping");
                    break;
                case "gotping":
                    //clearInterval(interval)
                    item = this.result[id];
                    if (!item) {
                        break;
                    }
                    // item.loss = json.loss;
                    item.ping = json.ping || 0;
                    /*
                                item = {
                                    "group": json.group,
                                    "remark": json.remarks,
                                    "loss": json.loss,
                                    "ping": json.ping,
                                    "speed": "0.00KB"
                                }
                                */
                    this.result[id] = item;
                    this.updateRowAsync(id, item);
                    break;
                case "startspeed":
                    //inverval=setInterval("app.loopevent("+id+",\"speed\")",300)
                    this.loopevent(id, "speed");
                    break;
                case "gotspeed":
                    //clearInterval(interval)
                    item = this.result[id];
                    item.speed = json.speed;
                    item.maxspeed = json.maxspeed;
                    item.traffic = this.getItemTraffic(item) + (Number(json.traffic) || 0);
                    this.result[id] = item;
                    this.syncTotalTrafficFromResult();
                    this.updateRowAsync(id, item);
                    break;
                case "picsaving":
                    this.$notify.info("保存结果图片中……");
                    break;
                case "picsaved":
                    this.$notify.success("图片已保存！路径：" + json.path);
                    break;
                case "picdata":
                    this.picdata = json.data;
                    break;
                case "eof":
                    this.loading = false;
                    this.syncTotalTrafficFromResult();
                    this.handleSmartRename(false)
                        .catch(() => {})
                        .finally(() => {
                            this.$notify.success(`${this.result.length}个节点测试完成`);
                        });
                    break;
                case "retest":
                    item = this.result[id];
                    this.$notify.error(
                        "节点 " + item.group + " - " + item.remark + " 第一次测试无速度，将重新测试。"
                    );
                    break;
                case "nospeed":
                    item = this.result[id];
                    this.$notify.error(
                        "节点 " + item.group + " - " + item.remark + " 第二次测试无速度。"
                    );
                    item.speed = "0.00B";
                    item.maxspeed = "0.00B";
                    this.result[id] = item;
                    this.updateRow(id, item);
                    break;
                case "error":
                    switch (json.reason) {
                        case "noconnection":
                            item = this.result[id];
                            item.ping = "0.00";
                            item.loss = "100.00%";
                            this.$notify.error(
                                "节点 " + item.group + " - " + item.remark + " 无法连接。"
                            );
                            this.result[id] = item;
                            this.updateRow(id, item);
                            break;
                        case "noresolve":
                            item = this.result[id];
                            item.ping = "0.00";
                            item.loss = "100.00%";
                            this.$notify.error(
                                "节点 " + item.group + " - " + item.remark + " 无法解析到 IP 地址。"
                            );
                            this.result[id] = item;
                            this.updateRow(id, item);
                            break;
                        case "nonodes":
                            this.$alert("找不到任何节点。请检查订阅链接。", "错误");
                            break;
                        case "invalidsub":
                            this.$alert("订阅获取异常。请检查订阅链接。", "错误");
                            this.terminate()
                            break;
                        case "norecoglink":
                            this.$alert("找不到任何链接。请检查提供的链接格式。", "错误");
                            break;
                        case "unhandled":
                            this.$alert("程序异常退出！", "错误");
                            break;
                    }
                    console.log("error:" + json.reason);
                    break;
            }
        },
        floatSort: function (obj1, obj2) {
            return parseFloat(obj1.ping) - parseFloat(obj2.ping);
        },
        speedSort: function (obj1, obj2) {
            const speed1 = isNaN(this.getSpeed(obj1.speed)) ? -1 : this.getSpeed(obj1.speed);
            const speed2 = isNaN(this.getSpeed(obj2.speed)) ? -1 : this.getSpeed(obj2.speed);
            return speed1 - speed2;
        },
        maxSpeedSort: function (obj1, obj2) {
            const speed1 = isNaN(this.getSpeed(obj1.maxspeed)) ? -1 : this.getSpeed(obj1.maxspeed);
            const speed2 = isNaN(this.getSpeed(obj2.maxspeed)) ? -1 : this.getSpeed(obj2.maxspeed);
            return speed1 - speed2;
        },
        filterPing: function (value, row) {
            return value === "available" ? row.ping > 0 : true;
        },
        filterAvgSpeed: function (value, row) {
            const speed = isNaN(this.getSpeed(row.speed)) ? -1 : this.getSpeed(row.speed);
            return speed >= value;
        },
        filterMaxSpeed: function (value, row) {
            const speed = isNaN(this.getSpeed(row.maxspeed)) ? -1 : this.getSpeed(row.maxspeed);
            return speed >= value;
        },
        filterProtocol: function (value, row) {
            if (value === "vmess") {
                return row.protocol.startsWith("vmess")
            }
            if (value === "vless") {
                return row.protocol.startsWith("vless")
            }
            if (value === "trojan") {
                return row.protocol.startsWith("trojan")
            }
            return value === row.protocol
        },
        checkSelectable: function (row, index) {
            return !!row.link && row.hasOwnProperty("id") && row.testing !== true
        },
    }

}

</script>

<style>

.ag-header-cell-label {
   justify-content: center;
}

</style>