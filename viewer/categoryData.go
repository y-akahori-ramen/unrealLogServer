package viewer

import (
	"regexp"
)

type CategoryData struct {
	Name       string
	Checked    bool
	Visibility bool
	Open       bool
	Children   []*CategoryData
}

func NewCategoryData(categoryName string) *CategoryData {
	return &CategoryData{Name: categoryName, Checked: true, Visibility: true, Open: false, Children: []*CategoryData{}}
}

var logCategory = regexp.MustCompile(`^Log.+`)

type CaregoryDataBuilder struct {
	parentCategories map[string]*CategoryData
	root             *CategoryData
}

func NewCaregoryDataBuilder() *CaregoryDataBuilder {
	return &CaregoryDataBuilder{}
}

func (c *CaregoryDataBuilder) getParentCategoryData(parentName string) *CategoryData {
	if data, ok := c.parentCategories[parentName]; ok {
		return data
	} else {
		data = &CategoryData{Name: parentName, Checked: false, Visibility: true, Open: false, Children: []*CategoryData{}}
		c.parentCategories[data.Name] = data
		c.root.Children = append(c.root.Children, data)
		return data
	}
}

func (c *CaregoryDataBuilder) CreateCategoryData(categoryNames []string) *CategoryData {
	c.parentCategories = make(map[string]*CategoryData)

	c.root = NewCategoryData("All")
	c.root.Open = true // 1階層目だけ最初から開けておく
	c.parentCategories["All"] = c.root

	for _, categoryName := range categoryNames {
		categoryData := NewCategoryData(categoryName)

		// Logから始まるカテゴリが多いためLog*の子要素としてViewerのカテゴリ一覧の可読性を上げる
		if logCategory.MatchString(categoryName) {
			parentCategoryData := c.getParentCategoryData("Log*")
			parentCategoryData.Children = append(parentCategoryData.Children, categoryData)
		} else {
			c.root.Children = append(c.root.Children, categoryData)
		}
	}

	return c.root
}
