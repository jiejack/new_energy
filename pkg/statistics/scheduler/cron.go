package scheduler

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInvalidCronExpression = errors.New("invalid cron expression")
	ErrInvalidField          = errors.New("invalid cron field")
	ErrUnsupportedCharacter  = errors.New("unsupported special character")
)

// CronField 表示Cron表达式的一个字段
type CronField struct {
	Name     string
	Min      int
	Max      int
	Value    string
	Values   map[int]bool
	HasStep  bool
	Step     int
	HasRange bool
	RangeMin int
	RangeMax int
}

// CronParser Cron表达式解析器
type CronParser struct {
	fields    []*CronField
	expression string
	location  *time.Location
}

// CronExpression 解析后的Cron表达式
type CronExpression struct {
	Second     *CronField
	Minute     *CronField
	Hour       *CronField
	Day        *CronField
	Month      *CronField
	Weekday    *CronField
	Year       *CronField // 可选年份字段
	Expression string
	Location   *time.Location
}

// 预定义表达式映射
var predefinedExpressions = map[string]string{
	"@yearly":   "0 0 0 1 1 *",
	"@annually": "0 0 0 1 1 *",
	"@monthly":  "0 0 0 1 * *",
	"@weekly":   "0 0 0 * * 0",
	"@daily":    "0 0 0 * * *",
	"@midnight": "0 0 0 * * *",
	"@hourly":   "0 0 * * * *",
	"@minutely": "0 * * * * *",
	"@secondly": "* * * * * *",
}

// NewCronParser 创建Cron解析器
func NewCronParser() *CronParser {
	return &CronParser{
		location: time.Local,
	}
}

// WithLocation 设置时区
func (p *CronParser) WithLocation(loc *time.Location) *CronParser {
	p.location = loc
	return p
}

// Parse 解析Cron表达式
// 支持格式: "秒 分 时 日 月 周" 或 "分 时 日 月 周"
func (p *CronParser) Parse(expression string) (*CronExpression, error) {
	// 处理预定义表达式
	if expr, ok := predefinedExpressions[strings.ToLower(expression)]; ok {
		expression = expr
	}

	// 处理 @every 格式
	if strings.HasPrefix(expression, "@every ") {
		return p.parseEveryExpression(expression)
	}

	// 分割字段
	fields := strings.Fields(expression)
	if len(fields) < 5 || len(fields) > 7 {
		return nil, fmt.Errorf("%w: expected 5-7 fields, got %d", ErrInvalidCronExpression, len(fields))
	}

	// 标准化为6字段格式（秒 分 时 日 月 周）
	if len(fields) == 5 {
		fields = append([]string{"0"}, fields...)
	}

	// 如果只有6个字段，添加年份字段（默认*）
	if len(fields) == 6 {
		fields = append(fields, "*")
	}

	expr := &CronExpression{
		Expression: expression,
		Location:   p.location,
	}

	var err error

	// 解析秒字段 (0-59)
	expr.Second, err = p.parseField("second", fields[0], 0, 59)
	if err != nil {
		return nil, fmt.Errorf("second field: %w", err)
	}

	// 解析分字段 (0-59)
	expr.Minute, err = p.parseField("minute", fields[1], 0, 59)
	if err != nil {
		return nil, fmt.Errorf("minute field: %w", err)
	}

	// 解析时字段 (0-23)
	expr.Hour, err = p.parseField("hour", fields[2], 0, 23)
	if err != nil {
		return nil, fmt.Errorf("hour field: %w", err)
	}

	// 解析日字段 (1-31)
	expr.Day, err = p.parseField("day", fields[3], 1, 31)
	if err != nil {
		return nil, fmt.Errorf("day field: %w", err)
	}

	// 解析月字段 (1-12)
	expr.Month, err = p.parseField("month", fields[4], 1, 12)
	if err != nil {
		return nil, fmt.Errorf("month field: %w", err)
	}

	// 解析周字段 (0-6, 0=周日)
	expr.Weekday, err = p.parseField("weekday", fields[5], 0, 6)
	if err != nil {
		return nil, fmt.Errorf("weekday field: %w", err)
	}

	// 解析年字段
	expr.Year, err = p.parseField("year", fields[6], 1970, 2099)
	if err != nil {
		return nil, fmt.Errorf("year field: %w", err)
	}

	return expr, nil
}

// parseEveryExpression 解析 @every 格式表达式
func (p *CronParser) parseEveryExpression(expression string) (*CronExpression, error) {
	durationStr := strings.TrimPrefix(expression, "@every ")
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid duration: %v", ErrInvalidCronExpression, err)
	}

	// 根据duration生成cron表达式
	if duration < time.Second {
		return nil, fmt.Errorf("%w: duration must be at least 1 second", ErrInvalidCronExpression)
	}

	seconds := int(duration.Seconds())
	if seconds%86400 == 0 {
		// 每天或更长
		days := seconds / 86400
		return p.Parse(fmt.Sprintf("0 0 0 */%d * *", days))
	} else if seconds%3600 == 0 {
		// 每小时或更长
		hours := seconds / 3600
		return p.Parse(fmt.Sprintf("0 0 */%d * * *", hours))
	} else if seconds%60 == 0 {
		// 每分钟或更长
		minutes := seconds / 60
		return p.Parse(fmt.Sprintf("0 */%d * * * *", minutes))
	}

	// 每秒
	return p.Parse(fmt.Sprintf("*/%d * * * * * *", seconds))
}

// parseField 解析单个字段
func (p *CronParser) parseField(name, value string, min, max int) (*CronField, error) {
	field := &CronField{
		Name:   name,
		Min:    min,
		Max:    max,
		Value:  value,
		Values: make(map[int]bool),
	}

	// 处理特殊字符
	if value == "*" || value == "?" {
		for i := min; i <= max; i++ {
			field.Values[i] = true
		}
		return field, nil
	}

	// 处理 L (最后)
	if value == "L" {
		if name == "day" {
			field.Values[max] = true
			return field, nil
		}
		return nil, fmt.Errorf("%w: L only valid for day field", ErrUnsupportedCharacter)
	}

	// 处理 W (工作日)
	if strings.Contains(value, "W") {
		return p.parseWorkdayField(field, value, min, max)
	}

	// 处理 # (第N个星期X)
	if strings.Contains(value, "#") {
		return p.parseNthWeekdayField(field, value)
	}

	// 处理逗号分隔的多个值
	parts := strings.Split(value, ",")
	for _, part := range parts {
		if err := p.parseFieldPart(field, part, min, max); err != nil {
			return nil, err
		}
	}

	return field, nil
}

// parseFieldPart 解析字段的一部分
func (p *CronParser) parseFieldPart(field *CronField, part string, min, max int) error {
	// 处理步长
	step := 1
	if strings.Contains(part, "/") {
		slashParts := strings.Split(part, "/")
		if len(slashParts) != 2 {
			return fmt.Errorf("%w: invalid step format", ErrInvalidField)
		}
		part = slashParts[0]
		stepValue, err := strconv.Atoi(slashParts[1])
		if err != nil {
			return fmt.Errorf("%w: invalid step value", ErrInvalidField)
		}
		if stepValue <= 0 {
			return fmt.Errorf("%w: step must be positive", ErrInvalidField)
		}
		step = stepValue
		field.HasStep = true
		field.Step = step
	}

	// 处理范围
	if part == "*" {
		for i := min; i <= max; i += step {
			field.Values[i] = true
		}
		return nil
	}

	// 处理 L 后缀 (最后一天)
	if strings.HasSuffix(part, "L") && field.Name == "day" {
		dayStr := strings.TrimSuffix(part, "L")
		day, err := strconv.Atoi(dayStr)
		if err != nil {
			return fmt.Errorf("%w: invalid L day value", ErrInvalidField)
		}
		// L表示该月最后一天
		field.Values[-day] = true // 负数表示从月末倒数
		return nil
	}

	// 处理范围表达式
	if strings.Contains(part, "-") {
		rangeParts := strings.Split(part, "-")
		if len(rangeParts) != 2 {
			return fmt.Errorf("%w: invalid range format", ErrInvalidField)
		}

		rangeMin, err := p.parseValue(rangeParts[0], min, max)
		if err != nil {
			return err
		}

		rangeMax, err := p.parseValue(rangeParts[1], min, max)
		if err != nil {
			return err
		}

		if rangeMin > rangeMax {
			return fmt.Errorf("%w: range min > max", ErrInvalidField)
		}

		field.HasRange = true
		field.RangeMin = rangeMin
		field.RangeMax = rangeMax

		for i := rangeMin; i <= rangeMax; i += step {
			field.Values[i] = true
		}
		return nil
	}

	// 处理单个值
	value, err := p.parseValue(part, min, max)
	if err != nil {
		return err
	}

	for i := value; i <= max; i += step {
		field.Values[i] = true
	}

	return nil
}

// parseValue 解析单个值
func (p *CronParser) parseValue(s string, min, max int) (int, error) {
	// 处理月份和星期的名称
	monthNames := map[string]int{
		"jan": 1, "feb": 2, "mar": 3, "apr": 4, "may": 5, "jun": 6,
		"jul": 7, "aug": 8, "sep": 9, "oct": 10, "nov": 11, "dec": 12,
	}
	weekdayNames := map[string]int{
		"sun": 0, "mon": 1, "tue": 2, "wed": 3, "thu": 4, "fri": 5, "sat": 6,
	}

	s = strings.ToLower(s)
	if v, ok := monthNames[s]; ok {
		return v, nil
	}
	if v, ok := weekdayNames[s]; ok {
		return v, nil
	}

	value, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("%w: invalid value '%s'", ErrInvalidField, s)
	}

	if value < min || value > max {
		return 0, fmt.Errorf("%w: value %d out of range [%d, %d]", ErrInvalidField, value, min, max)
	}

	return value, nil
}

// parseWorkdayField 解析工作日字段 (W)
func (p *CronParser) parseWorkdayField(field *CronField, value string, min, max int) (*CronField, error) {
	// 格式: 15W (最近的工作日)
	re := regexp.MustCompile(`^(\d+)W$`)
	matches := re.FindStringSubmatch(value)
	if len(matches) != 2 {
		return nil, fmt.Errorf("%w: invalid W format", ErrInvalidField)
	}

	day, err := strconv.Atoi(matches[1])
	if err != nil {
		return nil, fmt.Errorf("%w: invalid W day value", ErrInvalidField)
	}

	if day < min || day > max {
		return nil, fmt.Errorf("%w: W day out of range", ErrInvalidField)
	}

	// W 表示最近的工作日，这里先记录，实际计算在Next时处理
	field.Values[day] = true
	field.Values[-day] = true // 用负数标记这是W类型

	return field, nil
}

// parseNthWeekdayField 解析第N个星期X字段 (#)
func (p *CronParser) parseNthWeekdayField(field *CronField, value string) (*CronField, error) {
	// 格式: 6#3 (第三个星期五)
	re := regexp.MustCompile(`^(\d+)#(\d+)$`)
	matches := re.FindStringSubmatch(value)
	if len(matches) != 3 {
		return nil, fmt.Errorf("%w: invalid # format", ErrInvalidField)
	}

	weekday, err := strconv.Atoi(matches[1])
	if err != nil {
		return nil, fmt.Errorf("%w: invalid # weekday value", ErrInvalidField)
	}

	nth, err := strconv.Atoi(matches[2])
	if err != nil {
		return nil, fmt.Errorf("%w: invalid # nth value", ErrInvalidField)
	}

	if weekday < 0 || weekday > 6 {
		return nil, fmt.Errorf("%w: weekday must be 0-6", ErrInvalidField)
	}

	if nth < 1 || nth > 5 {
		return nil, fmt.Errorf("%w: nth must be 1-5", ErrInvalidField)
	}

	// 用特殊值存储，实际计算在Next时处理
	// 格式: weekday*10 + nth (如 6#3 = 63)
	field.Values[weekday*10+nth] = true

	return field, nil
}

// Next 计算下次执行时间
func (e *CronExpression) Next(from time.Time) time.Time {
	return e.nextTime(from, false)
}

// NextAfter 计算指定时间之后的下次执行时间
func (e *CronExpression) NextAfter(from time.Time) time.Time {
	return e.nextTime(from, true)
}

// nextTime 计算下次执行时间的核心逻辑
func (e *CronExpression) nextTime(from time.Time, inclusive bool) time.Time {
	// 转换到指定时区
	from = from.In(e.Location)

	// 如果不包含当前时间，从下一秒开始
	if !inclusive {
		from = from.Add(time.Second)
	}

	// 限制最大迭代次数，防止无限循环
	maxIterations := 366 * 24 * 60 * 60 // 一年的秒数
	iterations := 0

	for iterations < maxIterations {
		iterations++

		// 检查年份
		if !e.Year.Values[from.Year()] {
			from = time.Date(from.Year()+1, 1, 1, 0, 0, 0, 0, e.Location)
			continue
		}

		// 检查月份
		if !e.Month.Values[int(from.Month())] {
			from = time.Date(from.Year(), from.Month()+1, 1, 0, 0, 0, 0, e.Location)
			continue
		}

		// 检查日（需要考虑周和特殊日）
		if !e.matchDay(from) {
			from = time.Date(from.Year(), from.Month(), from.Day()+1, 0, 0, 0, 0, e.Location)
			continue
		}

		// 检查小时
		if !e.Hour.Values[from.Hour()] {
			from = time.Date(from.Year(), from.Month(), from.Day(), from.Hour()+1, 0, 0, 0, e.Location)
			continue
		}

		// 检查分钟
		if !e.Minute.Values[from.Minute()] {
			from = from.Add(time.Minute - time.Duration(from.Second())*time.Second)
			continue
		}

		// 检查秒
		if !e.Second.Values[from.Second()] {
			from = from.Add(time.Second)
			continue
		}

		return from
	}

	// 无法找到有效时间
	return time.Time{}
}

// matchDay 检查日期是否匹配
func (e *CronExpression) matchDay(t time.Time) bool {
	// 检查日字段
	dayMatch := false
	for day := range e.Day.Values {
		if day > 0 && day == t.Day() {
			dayMatch = true
			break
		}
		// 处理L（月末）
		if day < 0 && -day < 100 {
			// W类型，工作日检查
			actualDay := -day
			if actualDay == t.Day() && t.Weekday() >= 1 && t.Weekday() <= 5 {
				dayMatch = true
				break
			}
		}
	}

	// 检查周字段
	weekdayMatch := false
	for w := range e.Weekday.Values {
		if w >= 0 && w < 10 && int(t.Weekday()) == w {
			weekdayMatch = true
			break
		}
		// 处理 # 格式 (第N个星期X)
		if w >= 10 {
			weekday := w / 10
			nth := w % 10
			if int(t.Weekday()) == weekday && e.isNthWeekday(t, nth) {
				weekdayMatch = true
				break
			}
		}
	}

	// 如果日字段使用了?，则只检查周
	if e.Day.Value == "?" {
		return weekdayMatch
	}

	// 如果周字段使用了?，则只检查日
	if e.Weekday.Value == "?" {
		return dayMatch
	}

	// 两者都检查，满足其一即可
	return dayMatch || weekdayMatch
}

// isNthWeekday 检查是否是第N个星期X
func (e *CronExpression) isNthWeekday(t time.Time, nth int) bool {
	// 计算当前日期是该月第几个该星期
	day := t.Day()
	weekday := int(t.Weekday())

	// 找到该月第一个该星期
	firstDay := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
	firstWeekday := int(firstDay.Weekday())

	// 计算第一个该星期的日期
	firstTargetDay := 1 + (weekday - firstWeekday + 7) % 7
	if firstTargetDay < 1 {
		firstTargetDay += 7
	}

	// 计算当前日期是第几个
	currentNth := (day - firstTargetDay) / 7 + 1

	return currentNth == nth
}

// Prev 计算上次执行时间
func (e *CronExpression) Prev(from time.Time) time.Time {
	from = from.In(e.Location)
	from = from.Add(-time.Second)

	maxIterations := 366 * 24 * 60 * 60
	iterations := 0

	for iterations < maxIterations {
		iterations++

		// 检查年份
		if !e.Year.Values[from.Year()] {
			from = time.Date(from.Year(), 12, 31, 23, 59, 59, 0, e.Location)
			continue
		}

		// 检查月份
		if !e.Month.Values[int(from.Month())] {
			from = time.Date(from.Year(), from.Month(), 1, 23, 59, 59, 0, e.Location).Add(-time.Second)
			continue
		}

		// 检查日
		if !e.matchDay(from) {
			from = time.Date(from.Year(), from.Month(), from.Day(), 23, 59, 59, 0, e.Location).Add(-time.Second)
			continue
		}

		// 检查小时
		if !e.Hour.Values[from.Hour()] {
			from = from.Add(-time.Hour - time.Duration(from.Minute())*time.Minute - time.Duration(from.Second())*time.Second)
			continue
		}

		// 检查分钟
		if !e.Minute.Values[from.Minute()] {
			from = from.Add(-time.Minute - time.Duration(from.Second())*time.Second)
			continue
		}

		// 检查秒
		if !e.Second.Values[from.Second()] {
			from = from.Add(-time.Second)
			continue
		}

		return from
	}

	return time.Time{}
}

// GetNextN 获取接下来N次执行时间
func (e *CronExpression) GetNextN(from time.Time, n int) []time.Time {
	result := make([]time.Time, 0, n)
	current := from

	for i := 0; i < n; i++ {
		next := e.Next(current)
		if next.IsZero() {
			break
		}
		result = append(result, next)
		current = next
	}

	return result
}

// Validate 验证Cron表达式
func (e *CronExpression) Validate() error {
	if e.Second == nil || e.Minute == nil || e.Hour == nil ||
		e.Day == nil || e.Month == nil || e.Weekday == nil {
		return ErrInvalidCronExpression
	}

	// 检查是否有有效的执行时间
	next := e.Next(time.Now())
	if next.IsZero() {
		return fmt.Errorf("%w: no valid execution time found", ErrInvalidCronExpression)
	}

	return nil
}

// String 返回表达式字符串
func (e *CronExpression) String() string {
	return e.Expression
}

// GetSeconds 获取秒字段值列表
func (e *CronExpression) GetSeconds() []int {
	return e.getFieldValues(e.Second, 0, 59)
}

// GetMinutes 获取分字段值列表
func (e *CronExpression) GetMinutes() []int {
	return e.getFieldValues(e.Minute, 0, 59)
}

// GetHours 获取时字段值列表
func (e *CronExpression) GetHours() []int {
	return e.getFieldValues(e.Hour, 0, 23)
}

// GetDays 获取日字段值列表
func (e *CronExpression) GetDays() []int {
	return e.getFieldValues(e.Day, 1, 31)
}

// GetMonths 获取月字段值列表
func (e *CronExpression) GetMonths() []int {
	return e.getFieldValues(e.Month, 1, 12)
}

// GetWeekdays 获取周字段值列表
func (e *CronExpression) GetWeekdays() []int {
	return e.getFieldValues(e.Weekday, 0, 6)
}

// getFieldValues 获取字段值列表
func (e *CronExpression) getFieldValues(field *CronField, min, max int) []int {
	values := make([]int, 0)
	for i := min; i <= max; i++ {
		if field.Values[i] {
			values = append(values, i)
		}
	}
	return values
}

// CronParserBuilder Cron解析器构建器
type CronParserBuilder struct {
	parser *CronParser
}

// NewCronParserBuilder 创建Cron解析器构建器
func NewCronParserBuilder() *CronParserBuilder {
	return &CronParserBuilder{
		parser: NewCronParser(),
	}
}

// WithLocation 设置时区
func (b *CronParserBuilder) WithLocation(loc *time.Location) *CronParserBuilder {
	b.parser.location = loc
	return b
}

// Build 构建解析器
func (b *CronParserBuilder) Build() *CronParser {
	return b.parser
}

// MustParse 解析Cron表达式，出错时panic
func (p *CronParser) MustParse(expression string) *CronExpression {
	expr, err := p.Parse(expression)
	if err != nil {
		panic(err)
	}
	return expr
}

// IsValid 验证Cron表达式是否有效
func (p *CronParser) IsValid(expression string) bool {
	_, err := p.Parse(expression)
	return err == nil
}

// GetDescription 获取Cron表达式的描述
func (e *CronExpression) GetDescription() string {
	// 简单描述生成
	desc := "执行时间: "

	// 秒
	if e.Second.Value != "*" && e.Second.Value != "0" {
		desc += fmt.Sprintf("每分钟第%s秒 ", e.formatFieldValue(e.Second))
	}

	// 分
	if e.Minute.Value != "*" {
		desc += fmt.Sprintf("每小时第%s分 ", e.formatFieldValue(e.Minute))
	}

	// 时
	if e.Hour.Value != "*" {
		desc += fmt.Sprintf("每天%s点 ", e.formatFieldValue(e.Hour))
	}

	// 日
	if e.Day.Value != "*" && e.Day.Value != "?" {
		desc += fmt.Sprintf("每月%s日 ", e.formatFieldValue(e.Day))
	}

	// 月
	if e.Month.Value != "*" {
		desc += fmt.Sprintf("%s月 ", e.formatFieldValue(e.Month))
	}

	// 周
	if e.Weekday.Value != "*" && e.Weekday.Value != "?" {
		weekdays := []string{"周日", "周一", "周二", "周三", "周四", "周五", "周六"}
		desc += fmt.Sprintf("每%s ", weekdays[e.Weekday.Values[0]])
	}

	return strings.TrimSpace(desc)
}

// formatFieldValue 格式化字段值
func (e *CronExpression) formatFieldValue(field *CronField) string {
	values := make([]string, 0)
	for v := range field.Values {
		if v >= 0 && v < 100 {
			values = append(values, fmt.Sprintf("%d", v))
		}
	}
	if len(values) == 0 {
		return "*"
	}
	if len(values) == 1 {
		return values[0]
	}
	return strings.Join(values, ",")
}
