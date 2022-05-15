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

func (c *CaregoryDataBuilder) CreateCategoryData(categoryFilterInfos []FilterInfo) *CategoryData {
	c.parentCategories = make(map[string]*CategoryData)

	c.root = &CategoryData{Name: "All", Checked: false, Visibility: true, Open: true, Children: []*CategoryData{}}
	c.parentCategories["All"] = c.root

	for _, category := range categoryFilterInfos {
		categoryData := &CategoryData{Name: category.Name, Checked: category.Checked, Visibility: true, Open: false, Children: []*CategoryData{}}

		// Logから始まるカテゴリが多いためLog*の子要素としてViewerのカテゴリ一覧の可読性を上げる
		if logCategory.MatchString(category.Name) {
			parentCategoryData := c.getParentCategoryData("Log*")
			parentCategoryData.Children = append(parentCategoryData.Children, categoryData)
		} else {
			c.root.Children = append(c.root.Children, categoryData)
		}
	}

	return c.root
}
