<template>
  <div :class="className" :style="{height:height,width:width}" />
</template>

<script>
import echarts from 'echarts'
require('echarts/theme/macarons') // echarts theme
import resize from './mixins/resize'

export default {
  mixins: [resize],
  props: {
    className: {
      type: String,
      default: 'chart'
    },
    width: {
      type: String,
      default: '100%'
    },
    height: {
      type: String,
      default: '350px'
    },
    autoResize: {
      type: Boolean,
      default: true
    },
    chartData: {
      type: Object,
      required: true
    }
  },
  data() {
    return {
      chart: null
    }
  },
  watch: {
    chartData: {
      deep: true,
      handler(val) {
        this.setOptions(val)
      }
    }
  },
  mounted() {
    this.$nextTick(() => {
      this.initChart()
    })
  },
  beforeDestroy() {
    if (!this.chart) {
      return
    }
    this.chart.dispose()
    this.chart = null
  },
  methods: {
    initChart() {
      this.chart = echarts.init(this.$el, 'macarons')
      this.setOptions(this.chartData)
    },
    setOptions({ expectedData, actualData } = {}) {
      console.log('line', this.chartData)
      this.chart.setOption({
        xAxis: {
          data: ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'],
          boundaryGap: false,
          axisTick: {
            show: false
          }
        },
        grid: {
          left: 10,
          right: 10,
          bottom: 20,
          top: 30,
          containLabel: true
        },
        tooltip: {
          trigger: 'axis',
          axisPointer: {
            type: 'cross'
          },
          padding: [5, 10]
        },
        yAxis: {
          axisTick: {
            show: false
          }
        },
        legend: {
          // data: ['a.sechack-z.org', 'b.sechack-z.org']
          data: this.chartData.map(d => d.name)
        },
        series: this.chartData.map(d => {
          const color = (Math.random() * 0xFFFFFF | 0).toString(16)
          const randomColor = '#' + ('000000' + color).slice(-6)
          const animation = ['cubicInOut', 'quadraticOut']
          return {
            name: d.name,
            itemStyle: {
              normal: {
                color: randomColor,
                lineStyle: {
                  color: randomColor,
                  width: 2
                }
              }
            },
            smooth: true,
            type: 'line',
            data: d.count,
            animationDuration: 2800,
            animationEasing: animation[Math.floor(Math.random() * 2)]
          }
        })
        // series: [{
        // name: 'a.sechack-z.org', itemStyle: {
        //   normal: {
        //     color: '#FF005A',
        //     lineStyle: {
        //       color: '#FF005A',
        //       width: 2
        //     }
        //   }
        // },
        // smooth: true,
        // type: 'line',
        // data: expectedData,
        // animationDuration: 2800,
        // animationEasing: 'cubicInOut'
        // },
        // {
        //   name: 'b.sechack-z.org',
        //   smooth: true,
        //   type: 'line',
        //   itemStyle: {
        //     normal: {
        //       color: '#3888fa',
        //       lineStyle: {
        //         color: '#3888fa',
        //         width: 2
        //       },
        //       areaStyle: {
        //         color: '#f3f8ff'
        //       }
        //     }
        //   },
        //   data: actualData,
        //   animationDuration: 2800,
        //   animationEasing: 'quadraticOut'
        // }]
      })
    }
  }
}
</script>
